package goboot

import (
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// MustLoadConfig reads in a configuration files and environment variables in the following order
// of priority:
//
// 1. environment variables
// 2. {path}/config.{env}.yaml
// 3. {path}/config.yaml
//
// config.yaml is mandatory but config.{env}.yaml is not.
// An config variable "var.sub_2: value" can be overwritten with an environment variable VAR_SUB_2.
func MustLoadConfig(log zerolog.Logger, dir string, env string) *viper.Viper {
	v := viper.New()

	// Load {path}/config.yaml
	cfgDir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	mainCfg := cfgDir + "/config.yaml"
	v.SetConfigFile(mainCfg)

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	log.Info().Msgf("loaded configuration %q", mainCfg)

	// Load {path}/config.{env}.yaml
	config2 := cfgDir + "/config." + env + ".yaml"
	v.SetConfigFile(config2)

	// Load environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.MergeInConfig(); err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Info().Msgf("config file %q not found, skipping", config2)
		}
	} else {
		log.Info().Msgf("loaded configuration %q", config2)
	}

	// Workaround because viper does not treat env vars the same as other config:
	// https://github.com/spf13/viper/issues/188
	for _, key := range v.AllKeys() {
		val := v.Get(key)
		v.Set(key, val)
	}

	return v
}
