package zlog

import (
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap/zapcore"
)

func TestConfigBuildDefaults(t *testing.T) {
	logger := (Config{}).Build()
	if logger == nil {
		t.Fatal("Build returned a nil logger")
	}
	if logger.Core().Enabled(zapcore.DebugLevel) {
		t.Fatal("default logger should use info level")
	}
}

func TestZapWriterImplementsLogxWriter(t *testing.T) {
	var _ logx.Writer = NewZapWriter((Config{Level: "debug"}).Build())
}

func TestConfigLoadsExistingYAMLShape(t *testing.T) {
	var target struct {
		LoggerConfig Config
	}
	err := conf.LoadFromYamlBytes([]byte(`
LoggerConfig:
  level: debug
  stacktrace: true
  addCaller: true
  callerShip: 3
  mode: console
  json: true
  color: true
`), &target)
	if err != nil {
		t.Fatalf("load logger config: %v", err)
	}
	if target.LoggerConfig.Level != "debug" || target.LoggerConfig.Mode != ConsoleMode || !target.LoggerConfig.JSON {
		t.Fatalf("unexpected logger config: %+v", target.LoggerConfig)
	}
}
