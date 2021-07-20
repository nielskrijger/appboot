package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/connectivity"
)

func TestNewGrpcConnection_Success(t *testing.T) {
	conn := NewGrpcConnection(&ServiceConfig{
		Address: "test:50051",
	})

	assert.Equal(t, "test:50051", conn.Target())
	assert.Equal(t, connectivity.Idle, conn.GetState())
}

func TestNewGrpcConnection_WithTLSSuccess(t *testing.T) {
	conn := NewGrpcConnection(&ServiceConfig{
		Address: "test:50051",
		TLS: &TLSConfig{
			Enable: true,
		},
	})

	assert.Equal(t, "test:50051", conn.Target())
	assert.Equal(t, connectivity.Idle, conn.GetState())
}
