package config

import (
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/models"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/etcd"
	commonmysql "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/mysql"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/pulsar"
	logger "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/zlog"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQLConf        commonmysql.Conf
	RedisConf        redis.RedisConf
	WsConf           zrpc.RpcClientConf
	PulsarConfig     pulsar.PulsarConfig
	LoggerConfig     logger.Config
	Symbol           []models.Symbol
	Coin             []models.Coin
	EtcdRegisterConf etcd.EtcdRegisterConf `json:",optional"`
}

const (
	Ticker = "ticker"
	Tick   = "tick"
	Kline  = "kline"
	Depth  = "depth"
)
