package svc

import (
	"context"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/internal/config"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/internal/dao/mysqldao"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/defines"
	pulsarConfig "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/pulsar"
	logger "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/zlog"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/utils"
	"github.com/redis/go-redis/v9"
	"github.com/yitter/idgenerator-go/idgen"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config            config.Config
	MatchConsumerList []pulsar.Consumer
	JwtClient         *utils.JWT
	UserRepo          *mysqldao.UserRepo
	RedisCli          *redis.Client
	RelayerRedisCli   *redis.Client
	MatchProducers    map[string]pulsar.Producer
	MySQLCli          *gorm.DB
	OrderFinalRepo    *mysqldao.OrderFinalRepo
	MatchTradeRepo    *mysqldao.MatchTradeRepo
	SettleArchive     *mysqldao.SettleArchiveStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	logger.InitDefaultLogger(&c.LoggerConfig)
	logx.SetWriter(logger.NewZapWriter(logger.GetZapLogger()))
	logx.DisableStat()

	client, err := c.PulsarConfig.BuildClient()
	if err != nil {
		logx.Severef("init pulsar client failed %v", err)
	}
	consumers := make([]pulsar.Consumer, 0, 10)

	m := make(map[string]pulsar.Producer)
	for _, v := range c.Symbol {
		topic := pulsarConfig.Topic{
			Tenant:    pulsarConfig.PublicTenant,
			Namespace: pulsarConfig.ColumbinaNamespace,
			Topic:     defines.MatchTopicInputPrefix + v.Name,
		}
		producer, err := client.CreateProducer(pulsar.ProducerOptions{
			Topic:           topic.BuildTopic(),
			SendTimeout:     10 * time.Second,
			DisableBatching: true,
		})
		if err != nil {
			logx.Severef("create producer failed %v", err)
		}
		consumerTopic := pulsarConfig.Topic{
			Tenant:    pulsarConfig.PublicTenant,
			Namespace: pulsarConfig.ColumbinaNamespace,
			Topic:     defines.MatchTopicOutputPrefix + v.Name,
		}
		logx.Infof("create consumer topic %s", consumerTopic.BuildTopic())
		consumer, err := client.Subscribe(pulsar.ConsumerOptions{
			Topic:            consumerTopic.BuildTopic(),
			SubscriptionName: "account_" + v.Name,
		})
		if err != nil {
			logx.Severef("init match consumer error:%v", err)
			continue
		}
		consumers = append(consumers, consumer)

		m[v.Name] = producer
	}

	idgen.SetIdGenerator(idgen.NewIdGeneratorOptions(2))

	mysqlCli := c.MySQLConf.MustNewClient()
	if err := mysqldao.EnsureSchema(context.Background(), mysqlCli); err != nil {
		logx.Severef("ensure account mysql schema failed: %v", err)
	}
	orderFinalRepo := mysqldao.NewOrderFinalRepo(mysqlCli)
	matchTradeRepo := mysqldao.NewMatchTradeRepo(mysqlCli)
	userRepo := mysqldao.NewUserRepo(mysqlCli)

	sc := &ServiceContext{
		Config:            c,
		MatchConsumerList: consumers,
		JwtClient:         utils.NewJWT(&c.JwtConf),
		UserRepo:          userRepo,
		MatchProducers:    m,
		MySQLCli:          mysqlCli,
		OrderFinalRepo:    orderFinalRepo,
		MatchTradeRepo:    matchTradeRepo,
		SettleArchive:     mysqldao.NewSettleArchiveStore(mysqlCli),
		RedisCli: redis.NewClient(&redis.Options{
			Addr:         c.RedisConf.Host,
			Password:     c.RedisConf.Pass,
			PoolSize:     32,
			MinIdleConns: 4,
		}),
		RelayerRedisCli: redis.NewClient(&redis.Options{
			Addr:         c.RedisConf.Host,
			Password:     c.RedisConf.Pass,
			PoolSize:     16,
			MinIdleConns: 2,
		}),
	}
	return sc
}
