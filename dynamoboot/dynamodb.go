package dynamoboot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nielskrijger/goboot"
	"github.com/rs/zerolog"
)

var (
	defaultMigrationsTable = "migrations"

	errMissingConfig = errors.New("missing dynamodb configuration")
	errMissingRegion = errors.New("config \"dynamodb.region\" is required")
)

type DynamodbConfig struct {
	// The AWS region to connection to
	Region string `yaml:"region"`

	// When true connects to a DynamoDB instance on localhost:8000
	Local bool `yaml:"local"`

	// Name of the table keeping track of migration history
	MigrationsTable string `yaml:"migrationsTable"`
}

type DynamoDB struct {
	MigrationsTable string
	Migrations      []*Migration

	Client *dynamodb.Client
	Config *DynamodbConfig

	log zerolog.Logger
}

// Configure connects to DynamoDB.
func (db *DynamoDB) Configure(env *goboot.AppEnv) error {
	db.log = env.Log

	// unmarshal Config and set defaults
	db.Config = &DynamodbConfig{}

	if !env.Config.InConfig("dynamodb") {
		return errMissingConfig
	}

	if err := env.Config.Sub("dynamodb").Unmarshal(db.Config); err != nil {
		return fmt.Errorf("parsing dynamodb configuration: %w", err)
	}

	if !env.Config.IsSet("dynamodb.region") {
		return errMissingRegion
	}

	if err := env.Config.Sub("dynamodb").Unmarshal(db.Config); err != nil {
		return fmt.Errorf("parsing dynamodb configuration: %w", err)
	}

	if !env.Config.IsSet("dynamodb.migrationsTable") {
		db.Config.MigrationsTable = defaultMigrationsTable
	}

	if db.Config.Local {
		client, err := db.createLocalClient(context.Background())
		if err != nil {
			return fmt.Errorf("connecting to local dynamodb Client: %w", err)
		}

		db.Client = client
	} else {
		client, err := db.createClient(context.Background())
		if err != nil {
			return fmt.Errorf("creating dynamodb client: %w", err)
		}

		db.Client = client
	}

	// check if we can connect to DynamoDB
	if err := db.testConnectivity(context.Background()); err != nil {
		return err
	}

	return nil
}

// Name is needed for the AppService interface.
func (db *DynamoDB) Name() string {
	return "dynamodb"
}

// Init runs the DynamoDB migrations.
func (db *DynamoDB) Init() error {
	if err := db.Migrate(context.Background()); err != nil {
		return fmt.Errorf("running DynamoDB migrations: %w", err)
	}

	return nil
}

// Close is needed for the AppService interface.
func (db *DynamoDB) Close() error {
	return nil // DynamoDB Client does not need closing
}

// createLocalClient connects to a dynamodb in given region.
func (db *DynamoDB) createClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(db.Config.Region),
	)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}

// createLocalClient connects to a local dynamodb emulator on port 8000.
func (db *DynamoDB) createLocalClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(db.Config.Region), // value doesn't actually matter as long as it exists
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     "dummy",
				SecretAccessKey: "dummy",
				SessionToken:    "dummy",
				Source:          "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
	)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func (db *DynamoDB) testConnectivity(ctx context.Context) error {
	_, err := db.Client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return fmt.Errorf("connecting to DynamoDB: %w", err)
	}

	db.log.Info().Msg("successfully connected to DynamoDB")
	return nil
}

func (db *DynamoDB) TableExists(ctx context.Context, tableName string) (bool, error) {
	tables, err := db.Client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return false, fmt.Errorf("list tables: %w", err)
	}

	for _, n := range tables.TableNames {
		if n == tableName {
			return true, nil
		}
	}

	return false, nil
}

// CreateTable creates a table and waits until it is ready.
func (db *DynamoDB) CreateTable(ctx context.Context, tableInput *dynamodb.CreateTableInput) error {
	if _, err := db.Client.CreateTable(ctx, tableInput); err != nil {
		return fmt.Errorf("creating table %q: %w", *tableInput.TableName, err)
	}

	if err := db.waitForTable(ctx, *tableInput.TableName); err != nil {
		return err
	}

	db.log.Info().Msgf("created DynamoDB table %q", *tableInput.TableName)

	return nil
}

// CreateTableIfNotExists creates a table if it does not exist and waits until it is ready.
func (db *DynamoDB) CreateTableIfNotExists(ctx context.Context, tableInput *dynamodb.CreateTableInput) error {
	exists, err := db.TableExists(ctx, *tableInput.TableName)
	if err != nil {
		return err
	}

	if exists {
		db.log.Info().Msgf("table %q already exists, nothing to do here", *tableInput.TableName)
		return nil
	}

	return db.CreateTable(ctx, tableInput)
}

// waitForTable blocks until a DynamoDB table is ready for reading/writing.
func (db *DynamoDB) waitForTable(ctx context.Context, tableName string) error {
	w := dynamodb.NewTableExistsWaiter(db.Client)
	err := w.Wait(ctx,
		&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		},
		2*time.Minute,
		func(o *dynamodb.TableExistsWaiterOptions) {
			o.MaxDelay = 5 * time.Second
			o.MinDelay = 5 * time.Second
		})
	if err != nil {
		return fmt.Errorf("timed out while waiting for table to become active: %w", err)
	}

	return nil
}
