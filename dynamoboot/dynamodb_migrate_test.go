package dynamoboot_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nielskrijger/goboot/dynamoboot"
	"github.com/stretchr/testify/assert"
)

var testTable = aws.String("test")

type TestItem struct {
	ID   string
	Name string
}

var testMigrations = []*dynamoboot.Migration{
	{
		ID: "1",
		Migrate: func(db *dynamoboot.DynamoDB) error {
			return db.CreateTable(context.Background(), &dynamodb.CreateTableInput{ //nolint:wrapcheck
				TableName: testTable,
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
			})
		},
	},
	{
		ID: "2",
		Migrate: func(db *dynamoboot.DynamoDB) error {
			_, err := db.Client.PutItem(context.Background(), &dynamodb.PutItemInput{
				TableName: testTable,
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "1234"},
					"name": &types.AttributeValueMemberS{Value: "John Doe"},
				},
			})

			return err //nolint:wrapcheck
		},
	},
}

func TestDynamoDB_Migrate_Success(t *testing.T) {
	db := &dynamoboot.DynamoDB{Migrations: testMigrations}
	_ = setupDynamoDBEnv(t, db)

	// Run migrations
	assert.Nil(t, db.Init())

	// Verify inserted record
	res, err := db.Client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: testTable,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "1234"},
		},
	})
	assert.Nil(t, err)

	var item *TestItem
	err = attributevalue.UnmarshalMap(res.Item, &item)

	assert.Nil(t, err)
	assert.Equal(t, item.ID, "1234")
	assert.Equal(t, item.Name, "John Doe")
}

func TestDynamoDB_Migrate_InsertMigrationRecord(t *testing.T) {
	db := &dynamoboot.DynamoDB{Migrations: testMigrations}
	_ = setupDynamoDBEnv(t, db)

	// Run migrations
	assert.Nil(t, db.Init())

	// Verify migration records
	res, err := db.Client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(db.Config.MigrationsTable),
	})
	assert.Nil(t, err)

	var items []*dynamoboot.MigrationRecord
	err = attributevalue.UnmarshalListOfMaps(res.Items, &items)

	assert.Nil(t, err)
	assert.Equal(t, items[0].ID, "1")
	_, err = time.Parse("2006-01-02T15:04:05Z", items[0].Timestamp)
	assert.Nil(t, err)
	assert.Greater(t, items[0].Duration, int64(0))
}

func TestDynamoDBMigrate_RunOnce(t *testing.T) {
	runCount := 0

	db := &dynamoboot.DynamoDB{
		Migrations: []*dynamoboot.Migration{
			{
				ID: "1",
				Migrate: func(db *dynamoboot.DynamoDB) error {
					runCount++

					return nil
				},
			},
		},
	}
	_ = setupDynamoDBEnv(t, db)

	// Trigger migrations twice
	assert.Nil(t, db.Init())
	assert.Nil(t, db.Init())
	assert.Equal(t, 1, runCount)
}

func TestDynamoDBMigrate_ErrorWhenOutOfOrder(t *testing.T) {
	db := &dynamoboot.DynamoDB{Migrations: []*dynamoboot.Migration{testMigrations[0]}}
	_ = setupDynamoDBEnv(t, db)
	_ = db.Init() // Ensure first migration is inserted

	// Replace migration with another one with a different ID
	db.Migrations = []*dynamoboot.Migration{testMigrations[1]}
	err := db.Init()

	assert.EqualError(
		t,
		err,
		`running DynamoDB migrations: unexpected migration id "2", was expecting id "1" (you can only add new migrations at the end)`, //nolint:lll
	)
}

func TestDynamoDBMigrate_ErrorMigrationMissing(t *testing.T) {
	db := &dynamoboot.DynamoDB{Migrations: []*dynamoboot.Migration{testMigrations[0]}}
	_ = setupDynamoDBEnv(t, db)
	_ = db.Init() // Ensure first migration is inserted

	// Remove migration
	db.Migrations = []*dynamoboot.Migration{}
	err := db.Init()

	assert.EqualError(
		t,
		err,
		`running DynamoDB migrations: missing migration "1"; you're not allowed to delete migrations that have already run`, //nolint:lll
	)
}
