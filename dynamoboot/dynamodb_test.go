package dynamoboot_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/dynamoboot"
	"github.com/stretchr/testify/assert"
)

func setupDynamoDBEnv(t *testing.T, db *dynamoboot.DynamoDB) *goboot.AppEnv {
	t.Helper()

	env := goboot.NewAppEnv("./testdata", "valid")
	assert.Nil(t, db.Configure(env))

	_, _ = db.Client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
		TableName: aws.String(db.Config.MigrationsTable),
	})
	_, _ = db.Client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
		TableName: testTable,
	})

	return env
}

func TestDynamoDB_Success(t *testing.T) {
	s := &dynamoboot.DynamoDB{}

	env := setupDynamoDBEnv(t, s)
	defer func() { env.Close() }()

	assert.Nil(t, s.Configure(goboot.NewAppEnv("./testdata", "valid")))
	assert.Nil(t, s.Init())
	assert.NotNil(t, s.Client)
	assert.NotNil(t, s.Config)
}

func TestDynamoDB_ErrorMissingConfig(t *testing.T) {
	s := &dynamoboot.DynamoDB{}
	err := s.Configure(goboot.NewAppEnv("./testdata", ""))
	assert.EqualError(t, err, "missing dynamodb configuration")
}

func TestPostgres_ErrorMissingRegion(t *testing.T) {
	s := &dynamoboot.DynamoDB{}
	err := s.Configure(goboot.NewAppEnv("./testdata", "no-region"))
	assert.EqualError(t, err, "config \"dynamodb.region\" is required")
}
