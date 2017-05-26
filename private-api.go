package zaif

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encoding/json"

	"github.com/google/go-querystring/query"
)

const privateEndPointURL = "https://api.zaif.jp/tapi"

// PrivateAPI API有効にした際の キー,シークレットキー を設定
type PrivateAPI struct {
	Key        string
	Secret     string
	HTTPClient *http.Client
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
		Key:        key,
		Secret:     secret,
		HTTPClient: http.DefaultClient,
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

// Do API送信
func (api *PrivateAPI) Do(ctx context.Context, method string, param interface{}, out interface{}) error {
	v, err := query.Values(newTradingParam(method))
	if err != nil {
		return err
	}
	encodedParams := v.Encode()

	if param != nil {
		pv, err := query.Values(param)
		if err != nil {
			return err
		}
		encodedParams += "&" + pv.Encode()
	}

	req, err := http.NewRequest("POST", privateEndPointURL, strings.NewReader(encodedParams))
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Key", api.Key)
	req.Header.Add("Sign", makeSign(encodedParams, api.Secret))

	client := api.HTTPClient
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	return decoder.Decode(out)
}

type APIError struct {
	Message string
}

func (e APIError) Error() string {
	return e.Message
}

type apiResponse struct {
	Success int    `json:"success"`
	Error   string `json:"error"`
}

type GetInfoResponse struct {
	Deposit struct {
		BTC   float64 `json:"btc"`
		JPY   float64 `json:"jpy"`
		Kaori float64 `json:"kaori"`
		MONA  float64 `json:"mona"`
		XEM   float64 `json:"xem"`
	} `json:"deposit"`
	Funds struct {
		BTC   float64 `json:"btc"`
		JPY   float64 `json:"jpy"`
		Kaori float64 `json:"kaori"`
		MONA  float64 `json:"mona"`
		XEM   float64 `json:"xem"`
	} `json:"funds"`
	OpenOrders int `json:"open_orders"`
	Rights     struct {
		IDInfo       int `json:"id_info"`
		Info         int `json:"info"`
		PersonalInfo int `json:"personal_info"`
		Trade        int `json:"trade"`
		Withdraw     int `json:"withdraw"`
	} `json:"rights"`
	ServerTime int `json:"server_time"`
	TradeCount int `json:"trade_count"`
}

type getInfoAPIResponse struct {
	apiResponse
	Response *GetInfoResponse `json:"return"`
}

// GetInfo 現在の残高（余力および残高）、APIキーの権限、過去のトレード数、アクティブな注文数、サーバーのタイムスタンプを取得
func (api *PrivateAPI) GetInfo(ctx context.Context) (*GetInfoResponse, error) {
	var ret getInfoAPIResponse
	if err := api.Do(ctx, "get_info", nil, &ret); err != nil {
		return nil, err
	}
	if ret.Success == 0 {
		return nil, APIError{Message: ret.Error}
	}
	return ret.Response, nil
}

type ActiveOrdersResponse struct {
	ActiveOrders map[string]struct {
		CurrencyPair string  `json:"currency_pair"`
		Action       string  `json:"action"`
		Amount       float64 `json:"amount"`
		Price        int     `json:"price"`
		Timestamp    int     `json:"timestamp,string"`
	} `json:"active_orders"`
	TokenActiveOrders map[string]struct {
		CurrencyPair string  `json:"currency_pair"`
		Action       string  `json:"action"`
		Amount       float64 `json:"amount"`
		Price        int     `json:"price"`
		Timestamp    int     `json:"timestamp,string"`
	} `json:"token_active_orders"`
}

type activeOrdersAPIResponse struct {
	apiResponse
	Response *ActiveOrdersResponse `json:"return"`
}

// ActiveOrders 現在有効な注文一覧を取得
func (api *PrivateAPI) ActiveOrders(ctx context.Context, param ActiveOrdersRequest) (*ActiveOrdersResponse, error) {
	var ret activeOrdersAPIResponse
	if err := api.Do(ctx, "active_orders", param, &ret); err != nil {
		return nil, err
	}
	if ret.Success == 0 {
		return nil, APIError{Message: ret.Error}
	}
	return ret.Response, nil
}

type TradeResponse struct {
	Received float64 `json:"received"`
	Remains  int     `json:"remains"`
	OrderID  int     `json:"order_id"`
	Funds    struct {
		JPY  float64 `json:"jpy"`
		BTC  float64 `json:"btc"`
		MONA float64 `json:"mona"`
	} `json:"funds"`
}

type tradeAPIResponse struct {
	apiResponse
	Response *TradeResponse `json:"return"`
}

// Trade 注文
func (api *PrivateAPI) Trade(ctx context.Context, param TradeRequest) (*TradeResponse, error) {
	var ret tradeAPIResponse
	if err := api.Do(ctx, "trade", param, &ret); err != nil {
		return nil, err
	}
	if ret.Success == 0 {
		return nil, APIError{Message: ret.Error}
	}
	return ret.Response, nil
}

type CancelResponse struct {
	Funds struct {
		BTC   float64 `json:"btc"`
		JPY   float64 `json:"jpy"`
		Kaori float64 `json:"kaori"`
		MONA  float64 `json:"mona"`
	} `json:"funds"`
	OrderID int `json:"order_id"`
}

type cancelAPIResponse struct {
	apiResponse
	Response *CancelResponse `json:"return"`
}

// Cancel 注文キャンセル
func (api *PrivateAPI) Cancel(ctx context.Context, param CancelRequest) (*CancelResponse, error) {
	var ret cancelAPIResponse
	if err := api.Do(ctx, "cancel", param, &ret); err != nil {
		return nil, err
	}
	if ret.Success == 0 {
		return nil, APIError{Message: ret.Error}
	}
	return ret.Response, nil
}

type WithdrawResponse struct {
	Txid  string `json:"txid"`
	Funds struct {
		JPY  float64 `json:"jpy"`
		BTC  float64 `json:"btc"`
		XEM  float64 `json:"xem"`
		MONA float64 `json:"mona"`
	} `json:"funds"`
}

type withdrawAPIResponse struct {
	apiResponse
	Response *WithdrawResponse `json:"return"`
}

// Withdraw 資金の引き出しリクエストを送信
func (api *PrivateAPI) Withdraw(ctx context.Context, param WithdrawRequest) (*WithdrawResponse, error) {
	var ret withdrawAPIResponse
	if err := api.Do(ctx, "withdraw", param, &ret); err != nil {
		return nil, err
	}
	if ret.Success == 0 {
		return nil, APIError{Message: ret.Error}
	}
	return ret.Response, nil
}

type DepositHistoryResponse map[string]struct {
	Timestamp int     `json:"timestamp,string"`
	Address   string  `json:"address"`
	Amount    float64 `json:"amount"`
	Txid      string  `json:"txid"`
}

type depositHistoryAPIResponse struct {
	apiResponse
	Response *DepositHistoryResponse `json:"return"`
}

// DepositHistory 入金履歴を取得
func (api *PrivateAPI) DepositHistory(ctx context.Context, param DepositHistoryRequest) (*DepositHistoryResponse, error) {
	var ret depositHistoryAPIResponse
	if err := api.Do(ctx, "deposit_history", param, &ret); err != nil {
		return nil, err
	}
	if ret.Success == 0 {
		return nil, APIError{Message: ret.Error}
	}
	return ret.Response, nil
}

type WithdrawHistoryResponse struct {
	Timestamp int     `json:"timestamp,string"`
	Address   string  `json:"address"`
	Amount    float64 `json:"amount"`
	Fee       float64 `json:"fee"`
	Txid      string  `json:"txid"`
}

type withdrawHistoryAPIResponse struct {
	apiResponse
	Response map[string]WithdrawHistoryResponse `json:"return"`
}

// WithdrawHistory 出金履歴を取得
func (api *PrivateAPI) WithdrawHistory(ctx context.Context, param WithdrawHistoryRequest) (map[string]WithdrawHistoryResponse, error) {
	var ret withdrawHistoryAPIResponse
	if err := api.Do(ctx, "withdraw_history", param, &ret); err != nil {
		return nil, err
	}
	if ret.Success == 0 {
		return nil, APIError{Message: ret.Error}
	}
	return ret.Response, nil
}
