package goboot

import (
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/rs/zerolog"
)

func LogBulkStats(logger zerolog.Logger, bulk esutil.BulkIndexer, dur time.Duration) {
	if bulkStats := bulk.Stats(); bulkStats.NumFailed > 0 {
		logger.Error().Msgf(
			"finished indexing %s Elasticsearch documents with %s errors in %s (%s docs/sec)",
			humanize.Comma(int64(bulkStats.NumFlushed)),
			humanize.Comma(int64(bulkStats.NumFailed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(bulkStats.NumFlushed))),
		)
	} else {
		logger.Info().Msgf(
			"finished indexing %s Elasticsearch documents in %s (%s docs/sec)",
			humanize.Comma(int64(bulkStats.NumFlushed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(bulkStats.NumFlushed))),
		)
	}
}
