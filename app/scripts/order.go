package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
)

const defaultAPIBaseURL = "http://localhost:8888"

var (
	token      string
	apiBaseURL = strings.TrimRight(envOrDefault("COLUMBINA_API_BASE_URL", defaultAPIBaseURL), "/")
	symbolName = envOrDefault("COLUMBINA_SYMBOL_NAME", "COLU_USDT")
	symbolID   = int32EnvOrDefault("COLUMBINA_SYMBOL_ID", 1)
)

func main() {
	cli := resty.New()
	t := time.NewTimer(4 * time.Second)
	prices := getPrice()

	for {
		select {
		case price, ok := <-prices:
			if !ok {
				log.Fatal("价格数据连接已关闭")
			}
			if token == "" {
				continue
			}
			asksLevel, bidsLevel, err := getDepthLevel(cli)
			if err != nil {
				log.Printf("获取深度失败：%v", err)
				continue
			}
			if asksLevel < 15 {
				placeOrder(cli, price*1.03, 2)
			}
			if bidsLevel < 15 {
				placeOrder(cli, price*0.99, 1)
			}
		case <-t.C:
			login(cli)
		}
	}
}

func placeOrder(cli *resty.Client, price float64, side int32) {
	result := map[string]interface{}{}
	_, err := cli.R().
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"symbol_id":   symbolID,
			"symbol_name": symbolName,
			"price":       cast.ToString(price),
			"qty":         "1",
			"amount":      "",
			"side":        side,
			"order_type":  2,
		}).
		SetResult(&result).
		Post(apiURL("/order/v1/create_order"))
	if err != nil {
		log.Printf("下单失败：%v", err)
		return
	}
	if cast.ToInt64(result["code"]) != 0 {
		log.Printf("下单失败：%v", result)
		return
	}
	log.Printf("下单成功 price=%v side=%d", price, side)
}

func getDepthLevel(cli *resty.Client) (int, int, error) {
	result := map[string]interface{}{}
	_, err := cli.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{"symbol": symbolName, "level": 300}).
		SetResult(&result).
		Post(apiURL("/quotes/v1/get_depth_list"))
	if err != nil {
		return 0, 0, err
	}
	if cast.ToInt64(result["code"]) != 0 {
		return 0, 0, &apiError{operation: "获取深度", response: result}
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return 0, 0, &apiError{operation: "解析深度数据", response: result}
	}
	bids, bidsOK := data["bids"].([]interface{})
	asks, asksOK := data["asks"].([]interface{})
	if !bidsOK || !asksOK {
		return 0, 0, &apiError{operation: "解析深度档位", response: result}
	}
	return len(asks), len(bids), nil
}

func login(cli *resty.Client) {
	result := map[string]interface{}{}
	_, err := cli.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"username": envOrDefault("COLUMBINA_TEST_USERNAME", "zhangsan"),
			"password": envOrDefault("COLUMBINA_TEST_PASSWORD", "123456"),
		}).
		SetResult(&result).
		Post(apiURL("/account/v1/login"))
	if err != nil {
		log.Printf("登录失败：%v", err)
		return
	}
	if cast.ToInt64(result["code"]) != 0 {
		log.Printf("登录失败：%v", result)
		return
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		log.Printf("登录响应格式错误：%v", result)
		return
	}
	token = cast.ToString(data["token"])
}

type PriceData struct {
	C interface{} `json:"c"`
	P float64     `json:"p"`
	S string      `json:"s"`
	T int64       `json:"t"`
	V float64     `json:"v"`
}

type Resp struct {
	Data []*PriceData `json:"data"`
}

func getPrice() <-chan float64 {
	wsURL := strings.TrimSpace(os.Getenv("COLUMBINA_PRICE_WS_URL"))
	if wsURL == "" {
		log.Fatal("COLUMBINA_PRICE_WS_URL is required")
	}
	priceSymbol := envOrDefault("COLUMBINA_PRICE_SYMBOL", "BINANCE:COLUUSDT")

	prices := make(chan float64)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("连接价格 WebSocket 失败：%v", err)
	}
	subscription, err := json.Marshal(map[string]string{
		"type":   "subscribe",
		"symbol": priceSymbol,
	})
	if err != nil {
		log.Fatalf("生成价格订阅请求失败：%v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, subscription); err != nil {
		log.Fatalf("订阅价格失败：%v", err)
	}

	go func() {
		defer close(prices)
		defer conn.Close()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Printf("读取价格失败：%v", err)
				return
			}
			var resp Resp
			if err := json.Unmarshal(data, &resp); err != nil {
				log.Printf("解析价格失败：%v", err)
				continue
			}
			if len(resp.Data) > 0 {
				prices <- resp.Data[len(resp.Data)-1].P
			}
		}
	}()
	return prices
}

func apiURL(path string) string {
	return apiBaseURL + "/" + strings.TrimLeft(path, "/")
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func int32EnvOrDefault(key string, fallback int32) int32 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		log.Fatalf("%s 必须是整数：%v", key, err)
	}
	return int32(parsed)
}

type apiError struct {
	operation string
	response  map[string]interface{}
}

func (e *apiError) Error() string {
	data, _ := json.Marshal(e.response)
	return e.operation + "失败：" + string(data)
}
