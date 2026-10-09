package handler

import (
	"context"
	"sync"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/match/internal/engine"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/match/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/defines"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/models"
	pulsarConfig "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/pulsar"
	logger "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/zlog"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/proto/enum"
	matchMq "github.com/nmg988rbdv-svg/go-zero_columbina/common/proto/mq/match"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/utils"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

var (
	handlers sync.Map
)

func GetMatchHandler(symbol string) (*engine.MatchEngine, bool) {
	value, ok := handlers.Load(symbol)
	if !ok {
		return nil, false
	}
	return value.(*engine.MatchEngine), true
}

func InitMatchHandler(sc *svc.ServiceContext) {
	ctx := context.Background()

	for _, v := range sc.Config.Symbol {
		go func(symbol models.Symbol) {
			inputTopic := pulsarConfig.Topic{
				Tenant:    pulsarConfig.PublicTenant,
				Namespace: pulsarConfig.ColumbinaNamespace,
				Topic:     defines.MatchTopicInputPrefix + v.Name,
			}
			consumer, err := sc.PulsarClient.Subscribe(pulsar.ConsumerOptions{
				Topic:            inputTopic.BuildTopic(),
				SubscriptionName: "match_" + symbol.Name,
			})

			if err != nil {
				logx.Severef("init match handler error:%v", err)
			}
			outputTopic := pulsarConfig.Topic{
				Tenant:    pulsarConfig.PublicTenant,
				Namespace: pulsarConfig.ColumbinaNamespace,
				Topic:     defines.MatchTopicOutputPrefix + v.Name,
			}
			logx.Debugf("start match handler output topic=%s", outputTopic)
			producer, err := sc.PulsarClient.CreateProducer(pulsar.ProducerOptions{
				Topic: outputTopic.BuildTopic(),
			})
			if err != nil {
				logx.Severef("init pulsar producer failed %v", err)
			}
			me := engine.NewMatchEngine(symbol, sc.Config, producer, consumer, sc.RedisClient, sc.WsClient)
			me.Start()
			handlers.Store(symbol.Name, me)
			for {
				message, err := consumer.Receive(ctx)
				if err != nil {
					logx.Errorw("receive message fail", logger.ErrorField(err))
					continue
				}
				//message.ID().String()
				var matchReq matchMq.MatchInput
				if err := proto.Unmarshal(message.Payload(), &matchReq); err != nil {
					logx.Errorw("unmarshal message fail", logger.ErrorField(err))
					continue
				}

				logx.Infof("receive message data %v", &matchReq)
				var inputMessage *engine.InputMessage
				switch event := matchReq.Event.(type) {
				case *matchMq.MatchInput_CreateOrder:
					inputMessage = newCreateOrderInputMessage(&matchReq, event.CreateOrder, message.ID())

				case *matchMq.MatchInput_CancelOrder:
					inputMessage = &engine.InputMessage{
						MessageId:   matchReq.MessageId,
						PulsarMsgId: message.ID(),
						OrderPkId:   event.CancelOrder.Id,
						IsCancel:    true,
						Side:        event.CancelOrder.Side,
						OrderType:   event.CancelOrder.OrderType,
						Price:       utils.NewFromString(event.CancelOrder.Price),
					}

				}

				me.HandleOrder(inputMessage)
			}
		}(v)
	}

}

func newCreateOrderInputMessage(matchReq *matchMq.MatchInput, event *matchMq.CreateOrderEvent, messageID pulsar.MessageID) *engine.InputMessage {
	price := utils.NewFromString(event.Price)
	baseAmount := utils.NewFromString(event.BaseAmount)
	quoteAmount := utils.NewFromString(event.QuoteAmount)
	if event.OrderType == enum.OrderType_LO {
		// 限价单的委托总额始终由价格和数量计算，不能信任或依赖调用方传入值。
		// 卖单只冻结基础币，旧消息中的 QuoteAmount 可能为 0；撮合仍需要正确的委托总额。
		quoteAmount = price.Mul(baseAmount)
	}

	return &engine.InputMessage{
		MessageId:           matchReq.MessageId,
		PulsarMsgId:         messageID,
		Uid:                 event.Uid,
		OrderID:             event.OrderId,
		OrderPkId:           event.SequenceId,
		CreateTime:          0,
		IsCancel:            false,
		Price:               price,
		BaseAmount:          baseAmount,
		OrderType:           event.OrderType,
		QuoteAmount:         quoteAmount,
		Side:                event.Side,
		OrderStatus:         enum.OrderStatus_NewCreated,
		UnfilledBaseAmount:  baseAmount,
		FilledBaseAmount:    utils.DecimalZeroMaxPrec,
		UnfilledQuoteAmount: quoteAmount,
		FilledQuoteAmount:   utils.DecimalZeroMaxPrec,
	}
}
