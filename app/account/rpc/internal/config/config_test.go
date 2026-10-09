package config

import (
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestMySQLConfigLoads(t *testing.T) {
	for _, path := range []string{"../../etc/account.yaml", "../../etc/account.prod.yaml"} {
		var c Config
		if err := conf.Load(path, &c); err != nil {
			t.Fatalf("load %s: %v", path, err)
		}
		if c.MySQLConf.Host == "" || c.MySQLConf.Port == 0 || c.MySQLConf.Database != "trade" {
			t.Fatalf("invalid mysql config in %s: %+v", path, c.MySQLConf)
		}
	}
}
