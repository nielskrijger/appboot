package goboot

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// LoadConfig reads in configuration files and environment variables in the following order
// of priority:
//
// 1. environment variables (optional)
// 2. {path}/config.{env}.yaml (optional, but logs a warning if missing)
// 3. {path}/config.yaml (mandatory)
//
// An config variable "var.sub_2: value" can be overwritten with an environment variable VAR_SUB_2.
func LoadConfig(log zerolog.Logger, dir string, env string) (*viper.Viper, error) {
	v := viper.New()

	// Load {path}/config.yaml
	cfgDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("opening config dir %q: %w", dir, err)
	}

	mainCfg := cfgDir + "/config.yaml"
	v.SetConfigFile(mainCfg)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading config %q: %w", mainCfg, err)
	}

	log.Info().Msgf("loaded configuration %q", mainCfg)

	// Load {path}/config.{env}.yaml
	envCfg := cfgDir + "/config." + env + ".yaml"
	v.SetConfigFile(envCfg)

	// Load environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.MergeInConfig(); err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Warn().Err(err).Msgf("config file %q not found, skipping", envCfg)
		}
	} else {
		log.Info().Msgf("loaded configuration %q", envCfg)
	}

	// Viper ignores environment variables when unmarshalling if no defaults are set.
	// This should that, see also https://github.com/spf13/viper/issues/188
	for _, key := range v.AllKeys() {
		val := v.Get(key)
		v.Set(key, val)
	}

	return v, nil
}
