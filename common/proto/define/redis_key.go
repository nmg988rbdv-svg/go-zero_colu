package define

type RedisKey string

const (
	Ticker                   RedisKey = "columbina:ticker"
	Tick                     RedisKey = "columbina:tick"
	Kline                    RedisKey = "columbina:kline"
	AccountToken             RedisKey = "columbina:account:token"
	AccountSession           RedisKey = "columbina:account:session"
	AccountConsumedMessageId RedisKey = "columbina:account:consumed:messageId"
	OrderConsumedMessageId   RedisKey = "columbina:order:consumed:messageId"
	OpenOrder                RedisKey = "columbina:open_order"
	AccountMatchProcessed    RedisKey = "columbina:account_match_processed"
)

func (key RedisKey) WithSymbol(symbol string) string {
	return string(key) + "_" + symbol
}

func (key RedisKey) WithParams(params ...string) string {
	if len(params) == 0 {
		return string(key)
	}
	k := string(key)
	for _, v := range params {
		k += ":" + v
	}
	return k
}
