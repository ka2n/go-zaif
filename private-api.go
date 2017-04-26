package zaif

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

const privateEndPointURL = "https://api.zaif.jp/tapi"

// PrivateAPI API有効にした際の キー,シークレットキー を設定
type PrivateAPI struct {
	Key    string //
	Secret string //
}

// ActiveOrdersRequest Params of ActiveOrders
type ActiveOrdersRequest struct {
	From         int    `url:"from,omitempty"`
	Count        int    `url:"count,omitempty"`
	FromID       int    `url:"from_id,omitempty"`
	EndID        int    `url:"end_id,omitempty"`
	Order        string `url:"order,omitempty"`
	Since        int    `url:"since,omitempty"`
	End          int    `url:"end,omitempty"`
	CurrencyPair string `url:"currency_pair,omitempty"`
}

// TradeRequest Params of Trade
type TradeRequest struct {
	CurrencyPair string  `url:"currency_pair"`
	Action       string  `url:"action"`
	Price        int     `url:"price"`
	Amount       float32 `url:"amount"`
	Limit        bool    `url:"limit,omitempty"`
}

// CancelRequest Params of Cancel
type CancelRequest struct {
	OrderID int `url:"order_id"`
}

// WithdrawRequest Params of Withdraw
type WithdrawRequest struct {
	Currency string  `url:"currency"`
	Address  string  `url:"address"`
	Amount   float32 `url:"amount"`
	OptFee   float32 `url:"opt_fee,omitempty"`
}

// TradeHistoryRequest Params of TradeHistory
type TradeHistoryRequest struct {
	From         int    `url:"from,omitempty"`
	Count        int    `url:"count,omitempty"`
	FromID       int    `url:"from_id,omitempty"`
	EndID        int    `url:"end_id,omitempty"`
	Order        string `url:"order,omitempty"`
	Since        int    `url:"since,omitempty"`
	End          int    `url:"end,omitempty"`
	CurrencyPair string `url:"currency_pair,omitempty"`
}

// DepositHistoryRequest Params of DepositHistory()
type DepositHistoryRequest struct {
	From     int    `url:"from,omitempty"`
	Count    int    `url:"count,omitempty"`
	FromID   int    `url:"from_id,omitempty"`
	EndID    int    `url:"end_id,omitempty"`
	Order    string `url:"order,omitempty"`
	Since    int    `url:"since,omitempty"`
	End      int    `url:"end,omitempty"`
	Currency string `url:"currency"`
}

// WithdrawHistoryRequest Params of WithdrawHistory()
type WithdrawHistoryRequest struct {
	From     int    `url:"from,omitempty"`
	Count    int    `url:"count,omitempty"`
	FromID   int    `url:"from_id,omitempty"`
	EndID    int    `url:"end_id,omitempty"`
	Order    string `url:"order,omitempty"`
	Since    int    `url:"since,omitempty"`
	End      int    `url:"end,omitempty"`
	Currency string `url:"currency"`
}

// NewPrivateAPI To use PrivateAPI
func NewPrivateAPI(key string, secret string) *PrivateAPI {
	return &PrivateAPI{
		Key:    key,
		Secret: secret,
	}
}

// TradingParam 全リクエストで使用するparams
type TradingParam struct {
	Method string `url:"method"`
	Nonce  string `url:"nonce"`
}

// newTradingParam To make TradingParam
func newTradingParam(method string) TradingParam {
	return TradingParam{
		Method: method,
		Nonce:  newNonce(),
	}
}

// makeSign messageをHMAC-SHA512で署名
func makeSign(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha512.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// newNonce nonce取得
func newNonce() string {
	return fmt.Sprintf("%.6f", float64(time.Now().UnixNano())/float64(time.Second)-1420070400)
}

// Post API送信
func (api *PrivateAPI) Post(tradingParam TradingParam, parameter string) (string, error) {
	v, err := query.Values(tradingParam)
	if err != nil {
		return "", err
	}
	encodedParams := v.Encode()
	if parameter != "" {
		encodedParams += "&" + parameter
	}

	req, err := http.NewRequest("POST", privateEndPointURL, strings.NewReader(encodedParams))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", api.Key)
	req.Header.Add("Sign", makeSign(encodedParams, api.Secret))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetInfo 現在の残高（余力および残高）、APIキーの権限、過去のトレード数、アクティブな注文数、サーバーのタイムスタンプを取得
func (api *PrivateAPI) GetInfo() (string, error) {

	return api.Post(
		newTradingParam("get_info"),
		"",
	)
}

// ActiveOrders 現在有効な注文一覧を取得
func (api *PrivateAPI) ActiveOrders(body ActiveOrdersRequest) (string, error) {
	v, err := query.Values(body)
	if err != nil {
		return "", err
	}
	return api.Post(
		newTradingParam("active_orders"),
		v.Encode(),
	)
}

// Trade 注文
func (api *PrivateAPI) Trade(body TradeRequest) (string, error) {
	v, err := query.Values(body)
	if err != nil {
		return "", err
	}
	return api.Post(
		newTradingParam("trade"),
		v.Encode(),
	)
}

// Cancel 注文キャンセル
func (api *PrivateAPI) Cancel(body CancelRequest) (string, error) {
	v, err := query.Values(body)
	if err != nil {
		return "", err
	}
	return api.Post(
		newTradingParam("cancel"),
		v.Encode(),
	)
}

// Withdraw 資金の引き出しリクエストを送信
func (api *PrivateAPI) Withdraw(body WithdrawRequest) (string, error) {
	v, err := query.Values(body)
	if err != nil {
		return "", err
	}
	return api.Post(
		newTradingParam("withdraw"),
		v.Encode(),
	)
}

// DepositHistory 入金履歴を取得
func (api *PrivateAPI) DepositHistory(body DepositHistoryRequest) (string, error) {
	v, err := query.Values(body)
	if err != nil {
		return "", err
	}
	return api.Post(
		newTradingParam("deposit_history"),
		v.Encode(),
	)
}

// WithdrawHistory 出金履歴を取得
func (api *PrivateAPI) WithdrawHistory(body WithdrawHistoryRequest) (string, error) {
	v, err := query.Values(body)
	if err != nil {
		return "", err
	}
	return api.Post(
		newTradingParam("withdraw_history"),
		v.Encode(),
	)
}
