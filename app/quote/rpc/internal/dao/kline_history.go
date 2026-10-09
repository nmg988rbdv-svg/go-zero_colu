package dao

import (
	"context"
	"time"

	"github.com/nmg988rbdv-svg/go-zero_columbina/common/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KlineHistory struct {
	ID        uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Symbol    string `gorm:"column:symbol;size:32;uniqueIndex:uk_kline_symbol_type_start,priority:1;not null"`
	SymbolID  int32  `gorm:"column:symbol_id;not null"`
	KlineType int32  `gorm:"column:kline_type;uniqueIndex:uk_kline_symbol_type_start,priority:2;not null"`
	StartTime int64  `gorm:"column:start_time;uniqueIndex:uk_kline_symbol_type_start,priority:3;not null"`
	EndTime   int64  `gorm:"column:end_time;not null"`
	Open      string `gorm:"column:open_price;size:80;not null"`
	High      string `gorm:"column:high_price;size:80;not null"`
	Low       string `gorm:"column:low_price;size:80;not null"`
	Close     string `gorm:"column:close_price;size:80;not null"`
	Volume    string `gorm:"column:volume;size:80;not null"`
	Amount    string `gorm:"column:amount;size:80;not null"`
	Range     string `gorm:"column:price_range;size:80;not null"`
	UpdatedAt int64  `gorm:"column:updated_at;not null"`
}

func (KlineHistory) TableName() string { return "kline_history" }

type KlineHistoryRepo struct{ db *gorm.DB }

func NewKlineHistoryRepo(db *gorm.DB) *KlineHistoryRepo { return &KlineHistoryRepo{db: db} }

func EnsureSchema(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&KlineHistory{}, &TickDoc{})
}

func MemoryKlineToHistory(k *MemoryKline, symbol models.Symbol) *KlineHistory {
	if k == nil {
		return nil
	}
	return &KlineHistory{
		Symbol: symbol.Name, SymbolID: symbol.Id, KlineType: int32(k.KlineType),
		StartTime: k.StartTime, EndTime: k.EndTime, Open: k.Open.String(),
		High: k.High.String(), Low: k.Low.String(), Close: k.Close.String(),
		Volume: k.Volume.String(), Amount: k.Amount.String(), Range: k.Range,
		UpdatedAt: time.Now().UnixMilli(),
	}
}

func (r *KlineHistoryRepo) UpsertMany(ctx context.Context, rows []*KlineHistory) error {
	if len(rows) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	clean := make([]*KlineHistory, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if row.UpdatedAt == 0 {
			row.UpdatedAt = now
		}
		clean = append(clean, row)
	}
	if len(clean) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "symbol"}, {Name: "kline_type"}, {Name: "start_time"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"symbol_id", "end_time", "open_price", "high_price", "low_price",
			"close_price", "volume", "amount", "price_range", "updated_at",
		}),
	}).CreateInBatches(clean, 200).Error
}

func (r *KlineHistoryRepo) ListByRange(ctx context.Context, symbol string, klineType int32, startTime, endTime int64) ([]*KlineHistory, error) {
	db := r.db.WithContext(ctx).Where("symbol = ? AND kline_type = ?", symbol, klineType)
	if startTime > 0 {
		db = db.Where("start_time >= ?", startTime)
	}
	if endTime > 0 {
		db = db.Where("start_time <= ?", endTime)
	}
	var rows []*KlineHistory
	err := db.Order("start_time ASC").Find(&rows).Error
	return rows, err
}

func (r *KlineHistoryRepo) ListSince(ctx context.Context, symbol string, klineType int32, startTime int64) ([]*KlineHistory, error) {
	var rows []*KlineHistory
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND kline_type = ? AND start_time >= ?", symbol, klineType, startTime).
		Order("start_time ASC").Find(&rows).Error
	return rows, err
}

func HistoryToMemoryKline(d *KlineHistory) *MemoryKline {
	if d == nil {
		return nil
	}
	return &MemoryKline{
		KlineType: KlineType(d.KlineType), StartTime: d.StartTime, EndTime: d.EndTime,
		Open: mustDecimal(d.Open), High: mustDecimal(d.High), Low: mustDecimal(d.Low),
		Close: mustDecimal(d.Close), Volume: mustDecimal(d.Volume), Amount: mustDecimal(d.Amount), Range: d.Range,
	}
}
