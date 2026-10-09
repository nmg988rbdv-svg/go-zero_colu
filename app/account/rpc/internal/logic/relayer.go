package logic

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	relayerDiscoverInterval = time.Second
	relayerReadBlock        = time.Second
)

// StartRelayer 只消费实际存在的 Outbox Stream。
// 每个交易对仅使用一个阻塞读取连接，不再为 16384 个 Redis Slot 预创建 Stream
// 或启动数十个共享连接池的阻塞 worker。
func StartRelayer(rdb *redis.Client, pulsarProducers map[string]pulsar.Producer) {
	for symbol, producer := range pulsarProducers {
		go runSymbolRelayer(context.Background(), rdb, symbol, producer)
	}
}

func runSymbolRelayer(ctx context.Context, rdb *redis.Client, symbol string, producer pulsar.Producer) {
	group := fmt.Sprintf("relayer_group_%s", symbol)
	hostname, _ := os.Hostname()
	consumer := fmt.Sprintf("relayer_%s_%s", hostname, symbol)
	knownStreams := make(map[string]struct{})
	lastDiscover := time.Time{}

	for {
		if time.Since(lastDiscover) >= relayerDiscoverInterval {
			if err := discoverOutboxStreams(ctx, rdb, symbol, group, knownStreams); err != nil {
				logx.Errorf("discover redis outbox streams failed symbol=%s err=%v", symbol, err)
			}
			lastDiscover = time.Now()
		}

		if len(knownStreams) == 0 {
			if !waitRelayer(ctx, relayerDiscoverInterval) {
				return
			}
			continue
		}

		streams := make([]string, 0, len(knownStreams)*2)
		for stream := range knownStreams {
			streams = append(streams, stream)
		}
		sort.Strings(streams)
		streamCount := len(streams)
		for i := 0; i < streamCount; i++ {
			streams = append(streams, ">")
		}

		res, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  streams,
			Count:    10,
			Block:    relayerReadBlock,
		}).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if !errors.Is(err, redis.Nil) {
				logx.Errorf("read redis outbox failed symbol=%s err=%v", symbol, err)
				if !waitRelayer(ctx, time.Second) {
					return
				}
			}
			continue
		}

		for _, stream := range res {
			for _, message := range stream.Messages {
				processMessage(ctx, rdb, producer, stream.Stream, group, message)
			}
		}
	}
}

func discoverOutboxStreams(ctx context.Context, rdb *redis.Client, symbol, group string, known map[string]struct{}) error {
	pattern := fmt.Sprintf("mq_outbox_%s:{*}", symbol)
	iter := rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		stream := iter.Val()
		if _, ok := known[stream]; ok {
			continue
		}
		err := rdb.XGroupCreateMkStream(ctx, stream, group, "0").Err()
		if err != nil && !isGroupExistErr(err) {
			logx.Errorf("create redis outbox group failed stream=%s err=%v", stream, err)
			continue
		}
		known[stream] = struct{}{}
	}
	return iter.Err()
}

func waitRelayer(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func processMessage(ctx context.Context, rdb *redis.Client, producer pulsar.Producer, stream, group string, msg redis.XMessage) {
	value, ok := msg.Values["payload"]
	if !ok {
		rdb.XAck(ctx, stream, group, msg.ID)
		rdb.XDel(ctx, stream, msg.ID)
		return
	}
	payload, ok := value.(string)
	if !ok {
		logx.Errorf("invalid redis outbox payload stream=%s id=%s", stream, msg.ID)
		return
	}

	messageKey := msg.ID
	if value, ok := msg.Values["oid"].(string); ok {
		messageKey = value
	}

	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: []byte(payload),
		Key:     messageKey,
	}); err != nil {
		logx.Errorf("send redis outbox to pulsar failed stream=%s id=%s err=%v", stream, msg.ID, err)
		return
	}

	pipe := rdb.Pipeline()
	pipe.XAck(ctx, stream, group, msg.ID)
	pipe.XDel(ctx, stream, msg.ID)
	if _, err := pipe.Exec(ctx); err != nil {
		logx.Errorf("ack redis outbox failed stream=%s id=%s err=%v", stream, msg.ID, err)
	}
}

func isGroupExistErr(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "BUSYGROUP")
}
