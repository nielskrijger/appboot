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

	// Load environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Load {path}/config.{env}.yaml
	if env != "" {
		envCfg := cfgDir + "/config." + env + ".yaml"
		v.SetConfigFile(envCfg)

		if err := v.MergeInConfig(); err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				return nil, fmt.Errorf("config file not found %q: %w", envCfg, err)
			}

			return nil, fmt.Errorf("processing %q: %w", envCfg, err)
		}

		log.Info().Msgf("loaded configuration %q", envCfg)
	} else {
		log.Warn().Msg("environment variable ENV has not been set")
	}

	// Viper ignores environment variables when unmarshalling if no defaults are set.
	// This should fix that in some scenarios, see also https://github.com/spf13/viper/issues/188
	for _, key := range v.AllKeys() {
		val := v.Get(key)
		v.Set(key, val)
	}

	return v, nil
}
