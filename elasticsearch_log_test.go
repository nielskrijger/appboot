package goboot_test

import (
	"context"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goutils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var mockStats = esutil.BulkIndexerStats{
	NumAdded:    uint64(1),
	NumFlushed:  uint64(2),
	NumFailed:   uint64(0),
	NumIndexed:  uint64(3),
	NumCreated:  uint64(4),
	NumUpdated:  uint64(5),
	NumDeleted:  uint64(6),
	NumRequests: uint64(7),
}

type mockBulkIndexer struct {
	bulkStats esutil.BulkIndexerStats
}

func (m mockBulkIndexer) Add(context.Context, esutil.BulkIndexerItem) error {
	return nil
}

// Close waits until all added items are flushed and closes the indexer.
func (m mockBulkIndexer) Close(context.Context) error {
	return nil
}

// Stats returns indexer statistics.
func (m mockBulkIndexer) Stats() esutil.BulkIndexerStats {
	return m.bulkStats
}

func TestLogBulk_Success(t *testing.T) {
	log := &goutils.TestLogger{}
	bulkIndexer := mockBulkIndexer{bulkStats: mockStats}

	goboot.LogBulkStats(zerolog.New(log), bulkIndexer, 2*time.Second)

	assert.Equal(t, "Finished indexing 2 ElasticSearch documents in 2s (1 docs/sec)", log.LastLine()["message"])
}

func TestLogBulk_Failed(t *testing.T) {
	log := &goutils.TestLogger{}
	mockStatsClone := mockStats
	mockStatsClone.NumFailed = 3
	bulkIndexer := mockBulkIndexer{bulkStats: mockStatsClone}

	goboot.LogBulkStats(zerolog.New(log), bulkIndexer, 100*time.Millisecond)

	assert.Equal(t,
		"Finished indexing 2 ElasticSearch documents with 3 errors in 100ms (20 docs/sec)",
		log.LastLine()["message"],
	)
}
