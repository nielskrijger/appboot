package config_test

import (
	"os"
	"testing"

	"github.com/nielskrijger/goboot/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestConfig_LoadDefaultConfig(t *testing.T) {
	cfg := config.MustLoadConfig(zerolog.Nop(), "../testdata/conf", "unknown")
	assert.Equal(t, "config.yaml", cfg.GetString("vars.filename"))
	assert.Equal(t, "bar", cfg.GetString("vars.foo"))
	assert.Empty(t, cfg.GetString("vars.prod_only_var"))
}

func TestConfig_OverrideEnvConfig(t *testing.T) {
	cfg := config.MustLoadConfig(zerolog.Nop(), "../testdata/conf", "prod")
	assert.Equal(t, "config.prod.yaml", cfg.GetString("vars.filename"))
	assert.Equal(t, "bar", cfg.GetString("vars.foo"))
	assert.Equal(t, "config.prod.yaml", cfg.GetString("vars.prod_only_var"))
}

type TestConfig struct {
	Filename string `mapstructure:"filename"`
}

func TestConfig_OverrideEnvVariables(t *testing.T) {
	_ = os.Setenv("VARS_FILENAME", "from-env")
	_ = os.Setenv("VARS_PROD_ONLY_VAR", "from-env")
	cfg := config.MustLoadConfig(zerolog.Nop(), "../testdata/conf", "prod")
	assert.Equal(t, "from-env", cfg.GetString("vars.filename"))
	assert.Equal(t, "from-env", cfg.GetString("vars.prod_only_var"))

	// Viper ignores environment variables when unmarshalling, our utility
	// should correct that. See also https://github.com/spf13/viper/issues/188
	cfgStruct := &TestConfig{}
	err := cfg.Sub("vars").Unmarshal(cfgStruct)
	assert.Nil(t, err)
	assert.Equal(t, "from-env", cfgStruct.Filename)
}
