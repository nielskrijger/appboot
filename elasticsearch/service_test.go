package elasticsearch_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/elasticsearch"
	"github.com/stretchr/testify/assert"
)

func TestService_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &elasticsearch.Service{}
	assert.Nil(t, s.Configure(goboot.NewAppContext("../testdata/conf", "elasticsearch")))
	assert.Nil(t, s.Init())
}

func TestService_ErrorMissingConfig(t *testing.T) {
	s := &elasticsearch.Service{}
	err := s.Configure(goboot.NewAppContext("../testdata/conf", "empty"))
	assert.EqualError(t, err, "missing elasticsearch configuration")
}

func TestService_ErrorOnConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := &elasticsearch.Service{}
	err := s.Configure(goboot.NewAppContext("../testdata/conf", "elasticsearch-invalid"))
	assert.Contains(t, err.Error(), "expected 200 OK but got \"401 Unauthorized\" while retrieving elasticsearch info")
}
