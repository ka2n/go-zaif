package zaif

import (
	"golang.org/x/net/websocket"
)

type StreamResponse struct {
	Asks        [][]float64 `json:"asks"`
	Bids        [][]float64 `json:"bids"`
	TargetUsers []string    `json:"target_users"`
	Trades      []struct {
		CurrentyPair string  `json:"currenty_pair"`
		TradeType    string  `json:"trade_type"`
		Price        float64 `json:"price"`
		CurrencyPair string  `json:"currency_pair"`
		Tid          int64   `json:"tid"`
		Amount       float64 `json:"amount"`
		Date         int     `json:"date"`
	} `json:"trades"`
	LastPrice struct {
		Action string  `json:"action"`
		Price  float64 `json:"price"`
	} `json:"last_price"`
	CurrencyPair string `json:"currency_pair"`
	Timestamp    string `json:"timestamp"`
}

func (api TmpPublicAPI) Stream(pair string) (*websocket.Conn, error) {
	return websocket.Dial("wss://ws.zaif.jp:8888/stream?currency_pair="+pair, "", "http://localhost")
}

func (api TmpPublicAPI) ReceiveStream(conn *websocket.Conn) (*StreamResponse, error) {
	var res StreamResponse
	return &res, websocket.JSON.Receive(conn, &res)
}
