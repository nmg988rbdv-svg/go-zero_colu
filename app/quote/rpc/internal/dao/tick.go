package dao

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TickDoc struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	PkID        int64  `gorm:"column:pk_id;index;not null"`
	MatchID     string `gorm:"column:match_id;size:64;index;not null"`
	MatchSubID  string `gorm:"column:match_sub_id;size:64;uniqueIndex:uk_tick_match_role,priority:1;not null"`
	OrderID     string `gorm:"column:order_id;size:64;index;not null"`
	UserID      int64  `gorm:"column:user_id;index;not null"`
	Symbol      string `gorm:"column:symbol;size:32;index:idx_tick_symbol_created,priority:1;not null"`
	Price       string `gorm:"column:price;size:80;not null"`
	BaseAmount  string `gorm:"column:base_amount;size:80;not null"`
	QuoteAmount string `gorm:"column:quote_amount;size:80;not null"`
	Side        int32  `gorm:"column:side;not null"`
	Role        int32  `gorm:"column:role;uniqueIndex:uk_tick_match_role,priority:2;not null"`
	CreatedAt   int64  `gorm:"column:created_at;index:idx_tick_symbol_created,priority:2,sort:desc;not null"`
}

func (TickDoc) TableName() string { return "tick" }

type TickRepo struct{ db *gorm.DB }

func NewTickRepo(db *gorm.DB) *TickRepo { return &TickRepo{db: db} }

func (r *TickRepo) ListBySymbol(ctx context.Context, symbol string, limit int64) ([]*TickDoc, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []*TickDoc
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).
		Order("created_at DESC").Limit(int(limit)).Find(&rows).Error
	return rows, err
}

func (r *TickRepo) CountBySymbol(ctx context.Context, symbol string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TickDoc{}).Where("symbol = ?", symbol).Count(&count).Error
	return count, err
}

func (r *TickRepo) InsertMany(ctx context.Context, rows []*TickDoc) error {
	if len(rows) == 0 {
		return nil
	}
	clean := make([]*TickDoc, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			clean = append(clean, row)
		}
	}
	if len(clean) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(clean, 200).Error
}
