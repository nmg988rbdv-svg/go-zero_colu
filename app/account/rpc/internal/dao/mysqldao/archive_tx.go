package mysqldao

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettleArchiveStore struct{ db *gorm.DB }

func NewSettleArchiveStore(db *gorm.DB) *SettleArchiveStore { return &SettleArchiveStore{db: db} }

func EnsureSchema(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&User{}, &OrderFinal{}, &MatchTrade{})
}

func (s *SettleArchiveStore) InsertSettleData(ctx context.Context, trades []*MatchTrade, orders []*OrderFinal) error {
	if s == nil || s.db == nil || (len(trades) == 0 && len(orders) == 0) {
		return nil
	}
	normalizeMatchTrades(trades)
	normalizeOrderFinals(orders)
	trades = compactMatchTrades(trades)
	orders = compactOrderFinals(orders)

	var insertedTrades, insertedOrders int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(trades) > 0 {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(trades, 200)
			if result.Error != nil {
				return result.Error
			}
			insertedTrades = result.RowsAffected
		}
		if len(orders) > 0 {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(orders, 200)
			if result.Error != nil {
				return result.Error
			}
			insertedOrders = result.RowsAffected
		}
		return nil
	})
	if err != nil {
		logx.Errorw("mysql transaction insert settle data failed",
			logx.Field("tradeTotal", len(trades)), logx.Field("orderTotal", len(orders)), logx.Field("error", err.Error()))
		return fmt.Errorf("mysql transaction insert settle data failed: %w", err)
	}
	logx.Infow("mysql transaction insert settle data success",
		logx.Field("matchTradeInserted", insertedTrades),
		logx.Field("matchTradeSkipped", int64(len(trades))-insertedTrades),
		logx.Field("orderFinalInserted", insertedOrders),
		logx.Field("orderFinalSkipped", int64(len(orders))-insertedOrders))
	return nil
}

func compactMatchTrades(rows []*MatchTrade) []*MatchTrade {
	out := rows[:0]
	for _, row := range rows {
		if row != nil && row.MatchSubID != "" {
			out = append(out, row)
		}
	}
	return out
}

func compactOrderFinals(rows []*OrderFinal) []*OrderFinal {
	out := rows[:0]
	for _, row := range rows {
		if row != nil && row.OrderID != "" {
			out = append(out, row)
		}
	}
	return out
}

func normalizeMatchTrades(trades []*MatchTrade) {
	now := time.Now().UnixMilli()
	for _, trade := range trades {
		if trade != nil && trade.CreatedAt == 0 {
			trade.CreatedAt = now
		}
	}
}

func normalizeOrderFinals(orders []*OrderFinal) {
	now := time.Now().UnixMilli()
	for _, order := range orders {
		if order == nil {
			continue
		}
		if order.ArchivedAt == 0 {
			order.ArchivedAt = now
		}
		if order.FinishedAt == 0 {
			order.FinishedAt = now
		}
		if order.UpdatedAt == 0 {
			order.UpdatedAt = now
		}
	}
}
