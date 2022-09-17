package dynamoboot

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Migration struct {
	ID      string
	Migrate func(db *DynamoDB) error
}

type MigrationRecord struct {
	ID        string `dynamodbav:"id"`
	Timestamp string `dynamodbav:"timestamp"`
	Duration  int64  `dynamodbav:"duration"`
}

func (db *DynamoDB) Migrate(ctx context.Context) error {
	if err := db.CreateTableIfNotExists(ctx, db.createMigrationsTableInput()); err != nil {
		return err
	}

	if _, err := db.getMigrations(ctx); err != nil {
		return err
	}

	newMigrations, err := db.getNewMigrations(ctx)
	if err != nil {
		return err
	}

	if len(db.Migrations) == 0 {
		db.log.Info().Msg("DynamoDB is up-to-date")

		return nil
	}

	return db.runMigrations(ctx, newMigrations)
}

func (db *DynamoDB) createMigrationsTableInput() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: aws.String(db.Config.MigrationsTable),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("timestamp"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("timestamp"),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}
}

// getNewMigrations retrieves the migration history and returns all migrations
// that haven't run yet. Returns an error in the following scenarios:
//
// - The migration history has an unknown migration ID.
// - One of the new migrations has not been added to the back.
// - The migrations are ordered differently than the migration history.
func (db *DynamoDB) getNewMigrations(ctx context.Context) ([]*Migration, error) {
	records, err := db.getMigrations(ctx)
	if err != nil {
		return nil, err
	}

	var newMigrations []*Migration

	for i, migration := range db.Migrations {
		if i < len(records) {
			if migration.ID != records[i].ID {
				return nil, fmt.Errorf(
					"unexpected migration id %q, was expecting id %q (you can only add new migrations at the end)",
					migration.ID,
					records[i].ID,
				)
			}
		} else {
			newMigrations = append(newMigrations, migration)
		}
	}

	if len(records) > len(db.Migrations) {
		return nil, fmt.Errorf(
			"missing migration %q; you're not allowed to delete migrations that have already run",
			records[len(db.Migrations)].ID,
		)
	}

	return newMigrations, nil
}

// getMigrations retrieves all migrations that have run ordered by timestamp (oldest first).
func (db *DynamoDB) getMigrations(ctx context.Context) ([]MigrationRecord, error) {
	p := dynamodb.NewScanPaginator(db.Client, &dynamodb.ScanInput{
		TableName: aws.String(db.Config.MigrationsTable),
	})

	var items []MigrationRecord

	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetch DynamoDB page: %w", err)
		}

		var pItems []MigrationRecord

		err = attributevalue.UnmarshalListOfMaps(out.Items, &pItems)
		if err != nil {
			return nil, fmt.Errorf("unmarshal DynamoDB items: %w", err)
		}

		items = append(items, pItems...)
	}

	return items, nil
}

func (db *DynamoDB) runMigrations(ctx context.Context, migrations []*Migration) error {
	for _, migration := range migrations {
		start := time.Now()

		if err := migration.Migrate(db); err != nil {
			return fmt.Errorf("migration %q failed: %w", migration.ID, err)
		}

		elapsed := time.Since(start)
		if err := db.insertMigrationRecord(ctx, migration.ID, elapsed); err != nil {
			return err
		}
	}

	return nil
}

func (db *DynamoDB) insertMigrationRecord(ctx context.Context, id string, elapsed time.Duration) error {
	newRecord := MigrationRecord{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Duration:  elapsed.Milliseconds(),
	}

	av, err := attributevalue.MarshalMap(newRecord)
	if err != nil {
		return fmt.Errorf("marshal DynamoDB migration record: %w", err)
	}

	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.Config.MigrationsTable),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("insert DynamoDB migration record: %w", err)
	}

	return nil
}
