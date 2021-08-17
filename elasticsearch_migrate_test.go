//nolint:wrapcheck
package goboot_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/nielskrijger/goboot"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestElasticsearchMigrate_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Elasticsearch{
		Migrations: []*goboot.ElasticsearchMigration{
			{
				ID: "1",
				Migrate: func(es *goboot.Elasticsearch) error {
					return es.IndexCreate(context.Background(), "test")
				},
			},
			{
				ID: "2",
				Migrate: func(es *goboot.Elasticsearch) error {
					req := &esapi.IndexRequest{
						Index:      "test",
						DocumentID: "1",
						Body:       strings.NewReader(`{"foo": "bar"}`),
						Refresh:    "true",
					}
					_, err := req.Do(context.Background(), es.Client)

					return err
				},
			},
			{
				ID: "3",
				Migrate: func(es *goboot.Elasticsearch) error {
					req := &esapi.IndexRequest{
						Index:      "test",
						DocumentID: "2",
						Body:       strings.NewReader(`{"foo": "bar2"}`),
						Refresh:    "true",
					}
					_, err := req.Do(context.Background(), es.Client)

					return err
				},
			},
		},
	}
	setupElasticsearchEnv(t, s)

	// Run migrations
	assert.Nil(t, s.Init())

	req := esapi.SearchRequest{Index: []string{"test"}}
	res, err := req.Do(context.Background(), s.Client)
	assert.Nil(t, err)

	defer res.Body.Close()
	result, _ := io.ReadAll(res.Body)
	assert.Equal(t, `[{"foo": "bar"},{"foo": "bar2"}]`, gjson.GetBytes(result, "hits.hits.#._source").Raw)
}

func TestElasticsearchMigrate_RunOnce(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runCount := 0

	s := &goboot.Elasticsearch{
		Migrations: []*goboot.ElasticsearchMigration{
			{
				ID: "1",
				Migrate: func(es *goboot.Elasticsearch) error {
					runCount++

					return es.IndexCreate(context.Background(), "test") //nolint:wrapcheck
				},
			},
		},
	}
	setupElasticsearchEnv(t, s)

	// Run migrations twice
	assert.Nil(t, s.Init())
	assert.Nil(t, s.Init())

	assert.Equal(t, 1, runCount)
}

func TestElasticsearchMigrate_ErrorWhenOutOfOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Elasticsearch{
		Migrations: []*goboot.ElasticsearchMigration{
			{
				ID: "2",
				Migrate: func(es *goboot.Elasticsearch) error {
					return es.IndexCreate(context.Background(), "test") //nolint:wrapcheck
				},
			},
		},
	}
	setupElasticsearchEnv(t, s)

	// Add one migration in ES migrations index with a different id
	_ = s.InsertMigrationRecord("1", time.Millisecond)

	assert.EqualError(
		t,
		s.Init(),
		`running Elasticsearch migrations: unexpected migration id "2", was expecting id "1" (you can only add new migrations at the end)`, //nolint:lll
	)
}

func TestElasticsearchMigrate_ErrorMigrationMissing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &goboot.Elasticsearch{
		Migrations: []*goboot.ElasticsearchMigration{},
	}
	setupElasticsearchEnv(t, s)

	// Add one migration in ES migrations index with a different id
	_ = s.InsertMigrationRecord("1", time.Millisecond)

	assert.EqualError(
		t,
		s.Init(),
		`running Elasticsearch migrations: missing migration "1"; you're not allowed to delete migrations that have already run`, //nolint:lll
	)
}
