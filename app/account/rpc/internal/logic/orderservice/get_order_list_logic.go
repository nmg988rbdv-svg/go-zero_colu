package orderservicelogic

import (
	"context"
	"fmt"

	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/internal/dao/mysqldao"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/pb"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/proto/enum"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/rediskeys"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderListLogic {
	return &GetOrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetOrderList 当前委托从 Redis 查询；历史委托（全部成交/撤销/废弃）从 MySQL order_final 查询。
func (l *GetOrderListLogic) GetOrderList(in *pb.GetOrderListByUserReq) (*pb.GetOrderListByUserResp, error) {
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	activeStatuses, terminalStatuses := splitOrderStatuses(in.StatusList)
	if len(terminalStatuses) > 0 && len(activeStatuses) == 0 {
		return l.getHistoryOrderList(in, terminalStatuses, pageSize)
	}
	return l.getActiveOrderList(in, activeStatuses, pageSize)
}

func splitOrderStatuses(statusList []enum.OrderStatus) (active, terminal []enum.OrderStatus) {
	for _, s := range statusList {
		if mysqldao.IsTerminalStatus(s) {
			terminal = append(terminal, s)
		} else {
			active = append(active, s)
		}
	}
	return active, terminal
}

func (l *GetOrderListLogic) getHistoryOrderList(in *pb.GetOrderListByUserReq, statuses []enum.OrderStatus, pageSize int64) (*pb.GetOrderListByUserResp, error) {
	if l.svcCtx.OrderFinalRepo == nil {
		return nil, fmt.Errorf("order final repo not configured")
	}

	statusInts := make([]int32, 0, len(statuses))
	for _, s := range statuses {
		statusInts = append(statusInts, int32(s))
	}

	q := mysqldao.OrderFinalListQuery{
		UserID:     in.UserId,
		StatusList: statusInts,
		SymbolName: in.SymbolName,
		CursorID:   in.Id,
		PageSize:   pageSize,
	}

	total, err := l.svcCtx.OrderFinalRepo.CountByUser(l.ctx, q)
	if err != nil {
		return nil, fmt.Errorf("count history orders failed: %w", err)
	}

	docs, err := l.svcCtx.OrderFinalRepo.ListByUser(l.ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list history orders failed: %w", err)
	}

	orders := make([]*pb.Order, 0, len(docs))
	for _, doc := range docs {
		orders = append(orders, orderFinalToOrder(doc))
	}

	return &pb.GetOrderListByUserResp{
		OrderList: orders,
		Total:     total,
	}, nil
}

func (l *GetOrderListLogic) getActiveOrderList(in *pb.GetOrderListByUserReq, statuses []enum.OrderStatus, pageSize int64) (*pb.GetOrderListByUserResp, error) {
	tag := rediskeys.UserSlotTag(in.UserId)
	openOrdersKey := rediskeys.UserOpenOrdersKey(tag, in.UserId)
	activeOrdersKey := rediskeys.UserActiveOrdersKey(tag, in.UserId)

	statusSet := make(map[enum.OrderStatus]struct{}, len(statuses))
	for _, s := range statuses {
		statusSet[s] = struct{}{}
	}

	total, err := l.countMatchingOrders(openOrdersKey, activeOrdersKey, statusSet, in.SymbolName)
	if err != nil {
		return nil, err
	}

	orders, err := l.fetchPage(openOrdersKey, activeOrdersKey, statusSet, in.SymbolName, in.Id, pageSize)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrderListByUserResp{
		OrderList: orders,
		Total:     total,
	}, nil
}

func orderFinalToOrder(doc *mysqldao.OrderFinal) *pb.Order {
	if doc == nil {
		return nil
	}
	return &pb.Order{
		Id:                doc.PkID,
		OrderId:           doc.OrderID,
		UserId:            doc.UserID,
		SymbolId:          doc.SymbolID,
		SymbolName:        doc.SymbolName,
		BaseAmount:        doc.BaseAmount,
		Price:             doc.Price,
		QuoteAmount:       doc.QuoteAmount,
		Side:              enum.Side(doc.Side),
		Status:            enum.OrderStatus(doc.Status),
		OrderType:         enum.OrderType(doc.OrderType),
		FilledBaseAmount:  doc.FilledBaseAmount,
		FilledQuoteAmount: doc.FilledQuoteAmount,
		FilledAvgPrice:    doc.FilledAvgPrice,
		CreatedAt:         doc.CreatedAt,
		UpdatedAt:         doc.UpdatedAt,
	}
}

func (l *GetOrderListLogic) countMatchingOrders(openOrdersKey, activeOrdersKey string, statusSet map[enum.OrderStatus]struct{}, symbolName string) (int64, error) {
	if len(statusSet) == 0 && symbolName == "" {
		n, err := l.svcCtx.RedisCli.ZCard(l.ctx, openOrdersKey).Result()
		return n, err
	}

	orderIds, err := l.svcCtx.RedisCli.ZRevRange(l.ctx, openOrdersKey, 0, -1).Result()
	if err != nil {
		return 0, fmt.Errorf("zrevrange order index failed: %w", err)
	}
	return l.countFiltered(orderIds, activeOrdersKey, statusSet, symbolName)
}

func (l *GetOrderListLogic) countFiltered(orderIds []string, activeOrdersKey string, statusSet map[enum.OrderStatus]struct{}, symbolName string) (int64, error) {
	infos, err := l.loadOrderInfos(activeOrdersKey, orderIds)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, info := range infos {
		if matchesOrder(info, statusSet, symbolName) {
			total++
		}
	}
	return total, nil
}

func (l *GetOrderListLogic) fetchPage(openOrdersKey, activeOrdersKey string, statusSet map[enum.OrderStatus]struct{}, symbolName string, cursorId, pageSize int64) ([]*pb.Order, error) {
	max := "+inf"
	if cursorId > 0 {
		max = fmt.Sprintf("(%d", cursorId)
	}

	const maxRounds = 10
	orders := make([]*pb.Order, 0, pageSize)
	fetchLimit := pageSize * 2

	for round := 0; round < maxRounds && int64(len(orders)) < pageSize; round++ {
		orderIds, err := l.svcCtx.RedisCli.ZRevRangeByScore(l.ctx, openOrdersKey, &redis.ZRangeBy{
			Min:    "-inf",
			Max:    max,
			Offset: 0,
			Count:  fetchLimit,
		}).Result()
		if err != nil {
			return nil, fmt.Errorf("zrevrangebyscore order index failed: %w", err)
		}
		if len(orderIds) == 0 {
			break
		}

		infos, err := l.loadOrderInfos(activeOrdersKey, orderIds)
		if err != nil {
			return nil, err
		}
		for _, info := range infos {
			if !matchesOrder(info, statusSet, symbolName) {
				continue
			}
			orders = append(orders, orderInfoToOrder(info))
			if int64(len(orders)) >= pageSize {
				break
			}
		}

		if int64(len(orders)) >= pageSize {
			break
		}

		lastScore, err := l.svcCtx.RedisCli.ZScore(l.ctx, openOrdersKey, orderIds[len(orderIds)-1]).Result()
		if err != nil {
			return nil, fmt.Errorf("zscore order index failed: %w", err)
		}
		max = fmt.Sprintf("(%f", lastScore)
	}

	return orders, nil
}

func matchesOrder(info *pb.OrderInfo, statusSet map[enum.OrderStatus]struct{}, symbolName string) bool {
	if info == nil {
		return false
	}
	if len(statusSet) > 0 {
		if _, ok := statusSet[info.Status]; !ok {
			return false
		}
	}
	if symbolName != "" && info.SymbolName != symbolName {
		return false
	}
	return true
}

func (l *GetOrderListLogic) loadOrderInfos(activeOrdersKey string, orderIds []string) ([]*pb.OrderInfo, error) {
	if len(orderIds) == 0 {
		return nil, nil
	}
	values, err := l.svcCtx.RedisCli.HMGet(l.ctx, activeOrdersKey, orderIds...).Result()
	if err != nil {
		return nil, fmt.Errorf("hmget order info failed: %w", err)
	}

	infos := make([]*pb.OrderInfo, len(values))
	for i, value := range values {
		if value == nil {
			continue
		}
		var data []byte
		switch value := value.(type) {
		case string:
			data = []byte(value)
		case []byte:
			data = value
		default:
			l.Logger.Errorf("unexpected order info type %T", value)
			continue
		}
		var info pb.OrderInfo
		if err := proto.Unmarshal(data, &info); err != nil {
			l.Logger.Errorf("unmarshal order info failed: %v", err)
			continue
		}
		infos[i] = &info
	}
	return infos, nil
}

func orderInfoToOrder(info *pb.OrderInfo) *pb.Order {
	return &pb.Order{
		Id:                info.Id,
		OrderId:           info.OrderId,
		UserId:            info.UserId,
		SymbolId:          info.SymbolId,
		SymbolName:        info.SymbolName,
		BaseAmount:        info.BaseAmount,
		Price:             info.Price,
		QuoteAmount:       info.QuoteAmount,
		Side:              info.Side,
		Status:            info.Status,
		OrderType:         info.OrderType,
		FilledBaseAmount:  info.FilledBaseAmount,
		FilledQuoteAmount: info.FilledQuoteAmount,
		FilledAvgPrice:    info.FilledAvgPrice,
		CreatedAt:         info.CreatedAt,
		UpdatedAt:         info.UpdatedAt,
	}
}
