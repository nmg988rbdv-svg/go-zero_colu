package handler

import (
	"testing"

	"github.com/nmg988rbdv-svg/go-zero_columbina/common/proto/enum"
	matchMq "github.com/nmg988rbdv-svg/go-zero_columbina/common/proto/mq/match"
)

func TestNewCreateOrderInputMessageCalculatesLimitOrderQuoteAmount(t *testing.T) {
	request := &matchMq.MatchInput{MessageId: 42}
	event := &matchMq.CreateOrderEvent{
		OrderId:     "order-1",
		SequenceId:  7,
		Uid:         1,
		Side:        enum.Side_Sell,
		Price:       "1000",
		BaseAmount:  "100",
		QuoteAmount: "0",
		OrderType:   enum.OrderType_LO,
		SymbolName:  "COLU_USDT",
	}

	input := newCreateOrderInputMessage(request, event, nil)
	if got, want := input.QuoteAmount.String(), "100000"; got != want {
		t.Fatalf("quote amount = %s, want %s", got, want)
	}
	if got, want := input.UnfilledQuoteAmount.String(), "100000"; got != want {
		t.Fatalf("unfilled quote amount = %s, want %s", got, want)
	}
}
