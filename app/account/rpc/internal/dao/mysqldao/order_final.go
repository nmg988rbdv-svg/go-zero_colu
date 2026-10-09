package mysqldao

import (
	"context"

	"github.com/nmg988rbdv-svg/go-zero_columbina/common/proto/enum"
	"gorm.io/gorm"
)

type OrderFinal struct {
	ID                  uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID             string `gorm:"column:order_id;size:64;uniqueIndex:uk_order_final_order_id;not null"`
	UserID              int64  `gorm:"column:user_id;index:idx_order_final_user_pk,priority:1;not null"`
	PkID                int64  `gorm:"column:pk_id;index:idx_order_final_user_pk,priority:2,sort:desc;not null"`
	SymbolID            int32  `gorm:"column:symbol_id;not null"`
	SymbolName          string `gorm:"column:symbol_name;size:32;index:idx_order_final_symbol_name;not null"`
	Price               string `gorm:"column:price;size:80;not null"`
	BaseAmount          string `gorm:"column:base_amount;size:80;not null"`
	QuoteAmount         string `gorm:"column:quote_amount;size:80;not null"`
	Side                int32  `gorm:"column:side;not null"`
	Status              int32  `gorm:"column:status;index:idx_order_final_status;not null"`
	OrderType           int32  `gorm:"column:order_type;not null"`
	FilledBaseAmount    string `gorm:"column:filled_base_amount;size:80;not null"`
	UnFilledBaseAmount  string `gorm:"column:un_filled_base_amount;size:80;not null"`
	FilledQuoteAmount   string `gorm:"column:filled_quote_amount;size:80;not null"`
	UnFilledQuoteAmount string `gorm:"column:un_filled_quote_amount;size:80;not null"`
	FilledAvgPrice      string `gorm:"column:filled_avg_price;size:80;not null"`
	FinishReason        string `gorm:"column:finish_reason;size:32;not null"`
	CreatedAt           int64  `gorm:"column:created_at;not null"`
	UpdatedAt           int64  `gorm:"column:updated_at;not null"`
	FinishedAt          int64  `gorm:"column:finished_at;not null"`
	ArchivedAt          int64  `gorm:"column:archived_at;not null"`
}

func (OrderFinal) TableName() string { return "order_final" }

type OrderFinalRepo struct{ db *gorm.DB }

func NewOrderFinalRepo(db *gorm.DB) *OrderFinalRepo { return &OrderFinalRepo{db: db} }

type OrderFinalListQuery struct {
	UserID     int64
	StatusList []int32
	SymbolName string
	CursorID   int64
	PageSize   int64
}

func (r *OrderFinalRepo) query(ctx context.Context, q OrderFinalListQuery) *gorm.DB {
	db := r.db.WithContext(ctx).Model(&OrderFinal{}).Where("user_id = ?", q.UserID)
	if q.CursorID > 0 {
		db = db.Where("pk_id < ?", q.CursorID)
	}
	if len(q.StatusList) > 0 {
		db = db.Where("status IN ?", q.StatusList)
	}
	if q.SymbolName != "" {
		db = db.Where("symbol_name = ?", q.SymbolName)
	}
	return db
}

func (r *OrderFinalRepo) CountByUser(ctx context.Context, q OrderFinalListQuery) (int64, error) {
	var count int64
	err := r.query(ctx, q).Count(&count).Error
	return count, err
}

func (r *OrderFinalRepo) ListByUser(ctx context.Context, q OrderFinalListQuery) ([]*OrderFinal, error) {
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	var rows []*OrderFinal
	err := r.query(ctx, q).Order("pk_id DESC").Limit(int(pageSize)).Find(&rows).Error
	return rows, err
}

func IsTerminalStatus(status enum.OrderStatus) bool {
	switch status {
	case enum.OrderStatus_ALLFilled, enum.OrderStatus_Canceled, enum.OrderStatus_Wasted:
		return true
	default:
		return false
	}
}
