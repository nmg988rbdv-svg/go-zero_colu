package config

import (
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/zlog"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	LoggerConfig   zlog.Config
	AccountRpcConf zrpc.RpcClientConf
	MatchRpcConf   zrpc.RpcClientConf
	QuoteRpcConf   zrpc.RpcClientConf
	LangPath       string
}
