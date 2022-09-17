package goboot_test

import (
	"testing"

	"github.com/nielskrijger/goboot"
	"github.com/nielskrijger/goboot/test"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestConfig_LoadDefaultConfig(t *testing.T) {
	cfg, err := goboot.LoadConfig(zerolog.Nop(), "./testdata", "")
	assert.Nil(t, err)
	assert.Equal(t, "config.yaml", cfg.GetString("vars.filename"))
	assert.Equal(t, "bar", cfg.GetString("vars.foo"))
	assert.Empty(t, cfg.GetString("vars.prod_only_var"))
}

func TestConfig_OverrideEnvConfig(t *testing.T) {
	cfg, err := goboot.LoadConfig(zerolog.Nop(), "./testdata", "prod")
	assert.Nil(t, err)
	assert.Equal(t, "config.prod.yaml", cfg.GetString("vars.filename"))
	assert.Equal(t, "bar", cfg.GetString("vars.foo"))
	assert.Equal(t, "config.prod.yaml", cfg.GetString("vars.prod_only_var"))
}

func TestConfig_LogEmptyEnv(t *testing.T) {
	testLogger := &test.Logger{}

	_, err := goboot.LoadConfig(zerolog.New(testLogger), "./testdata", "")

	assert.Nil(t, err)
	assert.Equal(t, "environment variable ENV has not been set", testLogger.LastLine()["message"])
	assert.Equal(t, "warn", testLogger.LastLine()["level"])
}

func TestConfig_ErrorInvalidEnv(t *testing.T) {
	testLogger := &test.Logger{}

	_, err := goboot.LoadConfig(zerolog.New(testLogger), "./testdata", "unknown")

	assert.Contains(t, err.Error(), "config file not found")
	assert.Contains(t, err.Error(), "testdata/config.unknown.yaml")
}

type TestConfig struct {
	Filename string `mapstructure:"filename"`
}

func TestConfig_OverrideEnvVariables(t *testing.T) {
	t.Setenv("VARS_FILENAME", "from-env")
	t.Setenv("VARS_PROD_ONLY_VAR", "from-env")

	cfg, err := goboot.LoadConfig(zerolog.Nop(), "./testdata", "prod")
	assert.Nil(t, err)
	assert.Equal(t, "from-env", cfg.GetString("vars.filename"))
	assert.Equal(t, "from-env", cfg.GetString("vars.prod_only_var"))

	// Viper ignores environment variables when unmarshalling, our utility
	// should correct that. See also https://github.com/spf13/viper/issues/188
	cfgStruct := &TestConfig{}
	err = cfg.Sub("vars").Unmarshal(cfgStruct)
	assert.Nil(t, err)
	assert.Equal(t, "from-env", cfgStruct.Filename)
}
