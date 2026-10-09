package config

import (
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestConfigFilesLoad(t *testing.T) {
	for _, path := range []string{"../../etc/match.yaml", "../../etc/match.dev.yaml", "../../etc/match.prod.yaml"} {
		var c Config
		if err := conf.Load(path, &c); err != nil {
			t.Fatalf("load %s: %v", path, err)
		}
		if c.LoggerConfig.Level == "" || c.LoggerConfig.Mode == "" {
			t.Fatalf("invalid logger config in %s: %+v", path, c.LoggerConfig)
		}
		if c.RedisConf.Pass == "" {
			t.Fatalf("redis password is required in %s", path)
		}
	}
}
