package config

import (
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/models"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/etcd"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/pulsar"
	logger "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/zlog"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	PulsarConfig pulsar.PulsarConfig
	zrpc.RpcServerConf
	LoggerConfig     logger.Config
	Symbol           []models.Symbol
	Coin             []models.Coin
	WsConf           zrpc.RpcClientConf
	EtcdRegisterConf etcd.EtcdRegisterConf `json:",optional"`
	RedisConf        redis.RedisConf
}
