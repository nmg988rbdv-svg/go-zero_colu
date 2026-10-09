package config

import (
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/models"
	commonmysql "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/mysql"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/pulsar"
	logger "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/zlog"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/utils"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQLConf    commonmysql.Conf
	LoggerConfig logger.Config
	JwtConf      utils.JwtConf
	PulsarConfig pulsar.PulsarConfig
	RedisConf    redis.RedisConf
	Symbol       []models.Symbol
	Coin         []models.Coin
}
