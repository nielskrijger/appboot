package goboot

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrPubSubClosed = errors.New("pubsub Service has been closed")

// defaultDeadLetterName is the name used to identity the dead letter channel
// if no other name was defined.
const (
	DefaultDeadLetterName = "dead-letter"
	RetryDelay            = time.Minute * 2
	AckDeadline           = 10 * time.Second
	MaxAttributeLength    = 1024
)

// PubSub adds some utility methods to the Google cloud
// PubSub such ensuring a topic and subscription exists and
// deadlettering.
//
// It represents subscriptions and topics as a single message Channel
// as from an application perspective.
type PubSub struct {
	*pubsub.Client

	Channels map[string]*Channel

	// DeadLetter is the channel used for dead letter messages.
	DeadLetterChannel *Channel

	projectID string
	log       zerolog.Logger
	options   []Option
}

// PubSubRichMessage embeds the raw gcloud pubsub message with additional details
// and functions.
//
// The PubSubRichMessage primarily helps handling retryable and unrecoverable errors.
type RichMessage struct {
	*pubsub.Message
	Service *PubSub
	Channel *Channel
}

// PubSubChannel is a message channel containing a topic ID and optionally a subscription.
type Channel struct {
	ID             string
	TopicID        string
	SubscriptionID string

	// MaxRetryAge is the time since publishing the message within a recoverable error
	// is still NACK'ed rather than dead-lettered.
	//
	// The default MaxRetryAge is 2 minutes.
	//
	// The max age prevents messages from being requeued and retried thousands of times
	// until Google pubsub deletes them automatically after 7 days.
	//
	// When no dead letter channel is configured a message will always be NACK'ed upon a
	// recoverable error.
	MaxRetryAge time.Duration
}

type Option func(*PubSub)

// WithChannel option adds a channel with a topic and a subscription.
//
// The channel name is a self-chosen name separate from the topicID and subscriptionID
// to more easily reference the subscription in the rest of your codebase.
//
// If you're not intending to receive any messages you can leave the subscriptionID empty.
// Be aware any messages sent to a topic without any subscription are essentially lost.
func WithChannel(ch *Channel) func(*PubSub) {
	return func(cl *PubSub) {
		if ch.MaxRetryAge == 0 {
			ch.MaxRetryAge = RetryDelay
		}

		cl.addChannel(ch)
	}
}

// WithDeadLetter option adds a deadletter channel to the Pub/Sub service.
//
// The topic and optional subscription are automatically created if they don't exist
// already just like a regular channel.
//
// Without a dead letter channel messages will get NACKed on error and retried until
// Google pubsub automatically removes them after 7 days. This can quickly fill up
// your queues so you're highly advised to always add a dead letter channel.
//
// A RichMessage will get sent to the dead-letter channel if an unrecoverable error
// occurred or if the max message age has expired.
//
// Like a normal channel the subscriptionID is optional but be aware messages sent
// to a topic without any subscriptions are dropped immediately. When the channel
// name is left empty the default name "dead-letter" is used instead.
func WithDeadLetter(ch *Channel) func(*PubSub) {
	return func(cl *PubSub) {
		if ch.ID == "" {
			ch.ID = DefaultDeadLetterName
		}

		cl.addChannel(ch)
		cl.DeadLetterChannel = ch
	}
}

// NewPubSubService configures a new Service and connects to the pubsub server.
func NewPubSubService(projectID string, options ...Option) *PubSub {
	return &PubSub{
		projectID: projectID,
		Channels:  make(map[string]*Channel),
		options:   options,
	}
}

func (s *PubSub) Name() string {
	return "pubsub"
}

// Configure implements the context.AppService interface and instantiates
// the client connection to gcloud pubsub.
func (s *PubSub) Configure(appctx *AppContext) error {
	s.log = appctx.Log
	for _, option := range s.options {
		option(s)
	}

	client, err := pubsub.NewClient(context.Background(), s.projectID)
	if err != nil {
		return fmt.Errorf("connecting to gcloud pubsub: %w", err)
	}

	s.log.Info().Msgf("connected to %s pubsub", s.projectID)
	s.Client = client

	return nil
}

func (s *PubSub) addChannel(ch *Channel) {
	s.Channels[ch.ID] = ch
}

func (s *PubSub) Channel(channelID string) *Channel {
	return s.Channels[channelID]
}

// CreateAll ensures all topics and subscriptions exist.
func (s *PubSub) CreateAll() error {
	for _, ch := range s.Channels {
		if err := s.EnsureTopic(ch.TopicID); err != nil {
			return err
		}

		if ch.SubscriptionID != "" {
			if err := s.EnsureSubscription(ch.TopicID, ch.SubscriptionID); err != nil {
				return err
			}
		}
	}

	return nil
}

// Init implements the context.AppService interface and executes the MustCreateAll
// method.
func (s *PubSub) Init() error {
	s.log.Info().Msg("ensuring all google pubsub topics & subscriptions exist")

	return s.CreateAll()
}

// Close releases any resources held by the pubsub Service such as memory and goroutines.
func (s *PubSub) Close() error {
	if err := s.Client.Close(); err != nil {
		return fmt.Errorf("closing %s service: %w", s.Name(), err)
	}

	return nil
}

// DeadLetter publishes a copy of a message to the deadletter channel and ACK's
// the original message.
//
// If for some reason deadlettering the message failed an error is logged and the
// original message is NACK'ed.
//
// The dead letter message adds extra attributes to the original message.
//
// The method returns an error if neither neither ACKing or NACKing is possible.
func (msg *RichMessage) DeadLetter(ctx context.Context, cause error) error {
	if msg.Service.DeadLetterChannel == nil {
		return errors.New("no deadletter channel configured")
	}

	// Copy original msg attributes and add additional attributes
	newMap := make(map[string]string)
	for k, v := range msg.Attributes {
		newMap[k] = v
	}

	newMap["originalMessageID"] = msg.ID
	newMap["originalTopicID"] = msg.Channel.TopicID
	newMap["originalSubscriptionID"] = msg.Channel.SubscriptionID
	newMap["error"] = TrimLeftBytes(cause.Error(), MaxAttributeLength) // max attribute length is 1024 bytes

	if val, ok := newMap["deadLetterCount"]; ok {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil { //nolint:gomnd
			newMap["deadLetterCount"] = strconv.FormatInt(i+1, 10) //nolint:gomnd
		}
	} else {
		newMap["deadLetterCount"] = "1"
	}

	// Publish message to dead letter topic
	topic := msg.Service.Topic(msg.Service.DeadLetterChannel.TopicID)
	_, err := topic.Publish(ctx, &pubsub.Message{
		Data:       msg.Data,
		Attributes: newMap,
	}).Get(ctx)

	// When successful ACK, if unsuccessful NACK
	if err != nil {
		msg.Nack()

		return errors.Wrapf(err, "failed to sent message to dead letter topic %q", topic)
	}

	msg.Ack()

	return nil
}

// TryDeadLetter is the same as DeadLetter but logs any error rather than
// returning it.
//
// Messages will be redelivered automatically if not ACKed or NACKed in time.
func (msg *RichMessage) TryDeadLetter(ctx context.Context, cause error) {
	if err := msg.DeadLetter(ctx, cause); err != nil {
		msg.Service.log.Error().Err(err).Msg("failed to send message to dead letter queue")
	}
}

// RetryableError will NACK a message if it is within the max retry timespan,
// otherwise it will sent the message to a deadletter channel.
//
// Returns an error if no deadlettering the message failed.
func (msg *RichMessage) RetryableError(ctx context.Context, cause error) error {
	if time.Since(msg.PublishTime) > msg.Channel.MaxRetryAge {
		return msg.DeadLetter(ctx, cause)
	}

	// In all other cases NACK and let pubsub do a retry
	msg.Nack()

	return nil
}

// TryRetryableError is the same as RetryableError but logs any error rather than
// returning it.
//
// Messages will be redelivered automatically if not ACKed or NACKed in time.
func (msg *RichMessage) TryRetryableError(ctx context.Context, cause error) {
	if err := msg.RetryableError(ctx, cause); err != nil {
		msg.Service.log.Error().Err(err).Msg("failed processing retryable error")
	}
}

// EnsureTopic creates a topic with specified ID if it doesn't exist already.
// In most cases you should use CreateAll instead.
func (s *PubSub) EnsureTopic(topicID string) error {
	s.log.Info().Msgf("ensure topic %q exists", topicID)

	ctx := context.Background()
	exists, err := s.Topic(topicID).Exists(ctx)

	switch {
	case err != nil:
		return fmt.Errorf("checking if topic %s exists: %w", topicID, err)
	case !exists:
		if _, err := s.CreateTopic(ctx, topicID); err != nil {
			return fmt.Errorf("creating topic %s: %w", topicID, err)
		}

		s.log.Info().Msgf("created new topic %q", topicID)
	default:
		s.log.Info().Msgf("topic %q already exists", topicID)
	}

	return nil
}

// EnsureSubscription creates a subscription for specified topic. The topic
// must already exist.
//
// In most cases you should use CreateAll instead.
//
// The subscription is created with an ACK deadline of 10 seconds, meaning the
// message must be ACK'ed or NACK'ed within 10 seconds or else it will be re-delivered.
func (s *PubSub) EnsureSubscription(topicID string, subID string) error {
	s.log.Info().Msgf("ensure subscription %q for topic %q exists", subID, topicID)

	ctx := context.Background()
	exists, err := s.Subscription(subID).Exists(ctx)

	switch {
	case err != nil:
		return fmt.Errorf("checking if subscriptions %s exists: %w", subID, err)
	case !exists:
		_, err := s.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic:       s.Topic(topicID),
			AckDeadline: AckDeadline,
		})
		if err != nil {
			return fmt.Errorf("creating subscription %s: %w", subID, err)
		}

		s.log.Info().Msgf("created new subscription %q on topic %q", subID, topicID)
	default:
		s.log.Info().Msgf("subscription %q for topic %q already exists", subID, topicID)
	}

	return nil
}

// DeleteAll deletes all topics and subscriptions of all configured channels,
// including the dead-letter channel.
func (s *PubSub) DeleteAll() error {
	for _, ch := range s.Channels {
		if err := s.DeleteChannel(ch.ID); err != nil {
			return err
		}
	}

	return nil
}

// translateError returns a proper error message when the pubsub connection is
// closed.
//
// If the error was not a cancelled client connection the given error is wrapped
// with specified message.
func translateError(err error, wrapMsg string, args ...interface{}) error {
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() == codes.Canceled {
			return ErrPubSubClosed
		}

		return errors.Wrapf(err, wrapMsg, args...)
	}

	return nil
}

// DeleteChannel deletes the pubsub topic and subscription if they exist. If they don't exist
// nothing happens.
func (s *PubSub) DeleteChannel(channel string) error {
	ch := s.Channels[channel]
	if ch == nil {
		return errors.Errorf("channel %q not found", channel)
	}

	if ch.SubscriptionID != "" {
		ctx := context.Background()
		sub := s.Subscription(ch.SubscriptionID)

		if exists, err := sub.Exists(ctx); err != nil {
			return translateError(err, "failed to retrieve subscription %q", ch.SubscriptionID)
		} else if exists {
			if err := sub.Delete(ctx); err != nil {
				return translateError(err, "failed to delete subscription %q", ch.SubscriptionID)
			}
			s.log.Info().Msgf("deleted subscription %q on topic %q", ch.SubscriptionID, ch.TopicID)
		}
	}

	ctx := context.Background()
	topic := s.Topic(ch.TopicID)

	if exists, err := topic.Exists(ctx); err != nil {
		return translateError(err, "failed to retrieve topic %q", ch.TopicID)
	} else if exists {
		if err := topic.Delete(ctx); err != nil {
			return translateError(err, "failed to delete topic %q", ch.TopicID)
		}
		s.log.Info().Msgf("deleted topic %q", ch.TopicID)
	}

	return nil
}

// Receive starts receiving messages on specified channel.
//
// It is similar to a normal google pubsub subscription receiver but returns RichMessages
// in specified callback.
func (s *PubSub) Receive(ctx context.Context, channel string, f func(context.Context, *RichMessage)) error {
	ch := s.Channels[channel]
	if ch == nil {
		return errors.Errorf("channel %q not found", channel)
	}

	if ch.SubscriptionID == "" {
		return errors.Errorf("channel %q does not have a subscription", channel)
	}

	err := s.Subscription(ch.SubscriptionID).Receive(ctx, func(ctx2 context.Context, msg *pubsub.Message) {
		f(ctx2, &RichMessage{
			Message: msg,
			Service: s,
			Channel: ch,
		})
	})

	return translateError(err, "receiving message from subscription %q failed", ch.SubscriptionID)
}

// ReceiveNr blocks until the specified number of messages have been retrieved.
//
// This should only be used with caution for scripting and testing purposes.
func (s *PubSub) ReceiveNr(ctx context.Context, channel string, nrOfMessages int) ([]*RichMessage, error) {
	ch := s.Channels[channel]
	if ch == nil {
		return nil, errors.Errorf("channel %q not found", channel)
	}

	sub := s.Subscription(ch.SubscriptionID)
	cctx, cancel := context.WithCancel(ctx)

	var msgs []*RichMessage

	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		msgs = append(msgs, &RichMessage{
			Message: msg,
			Channel: ch,
			Service: s,
		})
		if len(msgs) >= nrOfMessages {
			cancel()
		}
	})
	if err != nil {
		return nil, translateError(err, "receiving message from subscription %q failed", ch.SubscriptionID)
	}

	return msgs, nil
}

// PublishEvent publishes a message to the channel's topic and waits for it to be published
// on the server.
//
// Google's pubsub batching is disabled by default which is only useful in very high-throughput
// use cases.
func (s *PubSub) PublishEvent(ctx context.Context, channel string, eventName string, payload interface{}) error {
	ch := s.Channels[channel]
	if ch == nil {
		return errors.Errorf("channel %q not found", channel)
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal payload for event %q on t %q", eventName, ch.TopicID)
	}

	t := s.Topic(ch.TopicID)

	_, err = t.Publish(ctx, &pubsub.Message{
		Data: bytes,
		Attributes: map[string]string{
			"event": eventName,
		},
	}).Get(ctx)
	if err != nil {
		return translateError(err, "could not publish event %q to t %q", eventName, ch.TopicID)
	}

	return nil
}

// TryPublishEvent is the same as PublishEvent but logs any error rather than
// returning it.
func (s *PubSub) TryPublishEvent(ctx context.Context, channel string, eventName string, payload interface{}) {
	if err := s.PublishEvent(ctx, channel, eventName, payload); err != nil {
		s.log.Error().Err(err).Msgf("failed to publish event %q", eventName)
	}
}

// TrimLeftBytes trims a string from the left until the string has max X bytes.
// Removes any invalid runes at the end.
func TrimLeftBytes(str string, maxBytes int) string {
	if len(str) < maxBytes {
		return str
	}

	// trim string, if it's valid ruturn it
	res := str[:maxBytes]
	if utf8.ValidString(res) {
		return res
	}

	// remove the the last invalid rune
	lastRune := maxBytes
	for lastRune > 0 && !utf8.RuneStart(str[lastRune]) {
		lastRune--
	}

	return res[:lastRune]
}
