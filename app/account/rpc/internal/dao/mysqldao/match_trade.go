package mysqldao

import "gorm.io/gorm"

type MatchTradeSide struct {
	PkID                int64  `gorm:"column:pk_id;not null"`
	UserID              int64  `gorm:"column:user_id;not null"`
	OrderID             string `gorm:"column:order_id;size:64;not null"`
	FilledBaseAmount    string `gorm:"column:filled_base_amount;size:80;not null"`
	UnFilledBaseAmount  string `gorm:"column:un_filled_base_amount;size:80;not null"`
	FilledQuoteAmount   string `gorm:"column:filled_quote_amount;size:80;not null"`
	UnFilledQuoteAmount string `gorm:"column:un_filled_quote_amount;size:80;not null"`
	OrderStatus         int32  `gorm:"column:order_status;not null"`
}

type MatchTrade struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	MatchSubID  string         `gorm:"column:match_sub_id;size:64;uniqueIndex:uk_match_trade_sub_id;not null"`
	MatchID     string         `gorm:"column:match_id;size:64;index:idx_match_trade_match_id;not null"`
	SymbolID    int32          `gorm:"column:symbol_id;not null"`
	SymbolName  string         `gorm:"column:symbol_name;size:32;index:idx_match_trade_symbol_name;not null"`
	BaseCoinID  int32          `gorm:"column:base_coin_id;not null"`
	QuoteCoinID int32          `gorm:"column:quote_coin_id;not null"`
	TakerIsBuy  bool           `gorm:"column:taker_is_buy;not null"`
	Price       string         `gorm:"column:price;size:80;not null"`
	BaseAmount  string         `gorm:"column:base_amount;size:80;not null"`
	QuoteAmount string         `gorm:"column:quote_amount;size:80;not null"`
	BeginPrice  string         `gorm:"column:begin_price;size:80;not null"`
	EndPrice    string         `gorm:"column:end_price;size:80;not null"`
	HighPrice   string         `gorm:"column:high_price;size:80;not null"`
	LowPrice    string         `gorm:"column:low_price;size:80;not null"`
	MatchTime   int64          `gorm:"column:match_time;index:idx_match_trade_match_time;not null"`
	Taker       MatchTradeSide `gorm:"embedded;embeddedPrefix:taker_"`
	Maker       MatchTradeSide `gorm:"embedded;embeddedPrefix:maker_"`
	CreatedAt   int64          `gorm:"column:created_at;not null"`
}

func (MatchTrade) TableName() string { return "match_trade" }

type MatchTradeRepo struct{ db *gorm.DB }

func NewMatchTradeRepo(db *gorm.DB) *MatchTradeRepo { return &MatchTradeRepo{db: db} }
