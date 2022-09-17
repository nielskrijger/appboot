package esboot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type Migration struct {
	ID      string
	Migrate func(es *Elasticsearch) error
}

type MigrationRecord struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Duration  string    `json:"duration"`
}

func (s *Elasticsearch) Migrate(ctx context.Context) error {
	exists, err := s.IndexExists(ctx, s.MigrationsIndex)
	if err != nil {
		return err
	}

	if !exists {
		s.log.Info().Msgf("elasticsearch %q index not found; run all migrations", s.MigrationsIndex)

		if err := s.IndexCreate(ctx, s.MigrationsIndex); err != nil {
			return err
		}
	}

	newMigrations, err := s.getNewMigrations(ctx)
	if err != nil {
		return err
	}

	if len(s.Migrations) == 0 {
		s.log.Info().Msg("Elasticsearch is up-to-date")

		return nil
	}

	return s.runMigrations(newMigrations)
}

// getNewMigrations retrieves the migration history and returns all migrations
// that haven't run yet. Returns an error in the following scenarios:
//
// - The migration history has an unknown migration ID.
// - One of the new migrations has not been added to the back.
// - The migrations are ordered differently than the migration history.
func (s *Elasticsearch) getNewMigrations(ctx context.Context) ([]*Migration, error) {
	var records []MigrationRecord
	if err := s.getMigrations(ctx, &records); err != nil {
		return nil, err
	}

	var newMigrations []*Migration

	for i, migration := range s.Migrations {
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

	if len(records) > len(s.Migrations) {
		return nil, fmt.Errorf(
			"missing migration %q; you're not allowed to delete migrations that have already run",
			records[len(s.Migrations)].ID,
		)
	}

	return newMigrations, nil
}

func (s *Elasticsearch) runMigrations(migrations []*Migration) error {
	for _, migration := range migrations {
		start := time.Now()

		if err := migration.Migrate(s); err != nil {
			return fmt.Errorf("migration %q failed: %w", migration.ID, err)
		}

		elapsed := time.Since(start)
		if err := s.InsertMigrationRecord(migration.ID, elapsed); err != nil {
			return err
		}
	}

	return nil
}

func (s *Elasticsearch) InsertMigrationRecord(id string, elapsed time.Duration) error {
	newRecord, err := json.Marshal(MigrationRecord{
		ID:        id,
		Timestamp: time.Now().UTC(),
		Duration:  elapsed.Truncate(time.Millisecond).String(),
	})
	if err != nil {
		return fmt.Errorf("marshal ES migration record: %w", err)
	}

	req := &esapi.IndexRequest{
		Index:      s.MigrationsIndex,
		DocumentID: id,
		Body:       bytes.NewReader(newRecord),
		Refresh:    "true",
	}

	if _, err = req.Do(context.Background(), s.Client); err != nil {
		return fmt.Errorf("insert ES migration record: %w", err)
	}

	return nil
}

func (s *Elasticsearch) IndexExists(ctx context.Context, idx string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{idx},
	}

	res, err := req.Do(ctx, s.Client)
	if err != nil {
		return false, fmt.Errorf("check if ES index %q exists: %w", idx, err)
	}

	return res.StatusCode == http.StatusOK, nil
}

func (s *Elasticsearch) IndexCreate(ctx context.Context, idx string) error {
	req := esapi.IndicesCreateRequest{Index: idx}

	res, err := req.Do(ctx, s.Client)
	if err != nil {
		return fmt.Errorf("creating ES index %q: %w", idx, err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status %d", res.StatusCode)
	}

	if err := s.ParseResponse(res, nil); err != nil {
		return err
	}

	s.log.Info().Msgf("created ES index %q", idx)

	return nil
}

func (s *Elasticsearch) IndexDelete(ctx context.Context, idx string) error {
	req := esapi.IndicesDeleteRequest{
		Index:             []string{idx},
		IgnoreUnavailable: esapi.BoolPtr(true),
	}

	res, err := req.Do(ctx, s.Client)
	if err != nil {
		return fmt.Errorf("deleting ES index %q: %w", idx, err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status %d", res.StatusCode)
	}

	if err := s.ParseResponse(res, nil); err != nil {
		return err
	}

	s.log.Info().Msgf("deleted ES index %q", idx)

	return nil
}

// getMigrations retrieves all migrations that have run.
func (s *Elasticsearch) getMigrations(ctx context.Context, r any) error {
	req := esapi.SearchRequest{
		Index: []string{s.MigrationsIndex},
	}

	res, err := req.Do(ctx, s.Client)
	if err != nil {
		return fmt.Errorf("search all ES documents in index %q: %w", s.MigrationsIndex, err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("index %q does not exist", res.StatusCode)
	}

	err = s.ParseResponse(res, &r)

	return err
}
