package goboot_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/mocks"
	"github.com/nielskrijger/goboot/utils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestAppContext_Logger(t *testing.T) {
	ctx := goboot.NewAppContext("./testdata", "")
	testLogger := &utils.TestLogger{}
	ctx.Log = zerolog.New(testLogger)

	ctx.Configure()

	entries := testLogger.Lines()
	assert.Len(t, entries, 2)
	assert.Equal(t, "starting configuring app services", entries[0]["message"])
	assert.Equal(t, "info", entries[1]["level"])
	assert.Equal(t, "finished configuring app services", entries[1]["message"])
	assert.Equal(t, "info", entries[1]["level"])
}

func TestAppContext_Configure(t *testing.T) {
	serviceMock1 := &mocks.AppService{}
	serviceMock2 := &mocks.AppService{}

	ctx := goboot.NewAppContext("./testdata", "")
	serviceMock1.On("Configure", ctx).Return(nil)
	serviceMock2.On("Configure", ctx).Return(nil)

	ctx.AddService(serviceMock1)
	ctx.AddService(serviceMock2)

	ctx.Configure()

	serviceMock1.AssertExpectations(t)
	serviceMock2.AssertExpectations(t)
}

func TestAppContext_Init(t *testing.T) {
	serviceMock1 := &mocks.AppService{}
	serviceMock1.On("Init").Return(nil)

	serviceMock2 := &mocks.AppService{}
	serviceMock2.On("Init").Return(nil)

	ctx := goboot.NewAppContext("./testdata", "")

	ctx.AddService(serviceMock1)
	ctx.AddService(serviceMock2)

	ctx.Init()

	serviceMock1.AssertExpectations(t)
	serviceMock2.AssertExpectations(t)
}

func TestAppContext_Close(t *testing.T) {
	serviceMock1 := &mocks.AppService{}
	serviceMock1.On("Close").Return(nil)

	serviceMock2 := &mocks.AppService{}
	serviceMock2.On("Close").Return(nil)

	ctx := goboot.NewAppContext("./testdata", "")
	ctx.AddService(serviceMock1)
	ctx.AddService(serviceMock2)

	ctx.Close()

	serviceMock1.AssertExpectations(t)
	serviceMock2.AssertExpectations(t)
}
