package zaif

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/net/websocket"
	"golang.org/x/sync/errgroup"
)

// StreamResponse stream API response
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

// NewStream stream API用のクライアントを作る
func NewStream() Stream {
	return Stream{
		subscriptions: make(map[string]chan *StreamResponse),
		connections:   make(map[string]*websocket.Conn),
	}
}

// Stream client
type Stream struct {
	subscriptions map[string]chan *StreamResponse
	connections   map[string]*websocket.Conn
	connected     bool
	mu            sync.Mutex

	Error error
}

// AddSubscription 指定ペアのsubscribe
// resはClose()が呼ばれた時にcloseされます
func (s *Stream) AddSubscription(pair string, res chan *StreamResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions[pair] = res
	return nil
}

// Receive stream APIからデータを受信開始する
func (s *Stream) Receive(ctx context.Context) error {
	// Make connections
	s.mu.Lock()
	if s.connected {
		return errors.New("already started")
	}
	s.connected = true
	for k := range s.subscriptions {
		conn, err := websocket.Dial("wss://ws.zaif.jp:8888/stream?currency_pair="+k, "", "http://localhost")
		if err != nil {
			for _, conn := range s.connections {
				conn.Close()
			}
			for _, sub := range s.subscriptions {
				close(sub)
			}
			s.mu.Unlock()
			return err
		}
		s.connections[k] = conn
	}

	// Receiving responses
	wg, ctx := errgroup.WithContext(ctx)
	for pair, conn := range s.connections {
		conn := conn
		c := s.subscriptions[pair]

		wg.Go(func() error {
			for {
				// if context already done, finish
				if ctx.Err() != nil {
					return nil
				}

				// Wait data...
				var res StreamResponse
				if err := websocket.JSON.Receive(conn, &res); err != nil {
					// this is not an error actually
					if ctx.Err() != nil {
						return nil
					}
					return err
				}

				// Publish to subscribers
				s.mu.Lock()
				if s.connected {
					c <- &res
				}
				s.mu.Unlock()
			}
		})
	}

	// Wait context to done
	// cleanup connections
	wg.Go(func() error {
		<-ctx.Done()
		s.mu.Lock()
		defer s.mu.Unlock()

		for _, sb := range s.connections {
			sb.Close()
		}
		s.connections = map[string]*websocket.Conn{}
		return nil
	})

	s.mu.Unlock()
	return wg.Wait()
}

// Close connections
func (s *Stream) Close() error {
	var err error
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, c := range s.subscriptions {
		close(c)
		s.subscriptions[k] = nil
		s.connections[k] = nil
	}
	s.connected = false
	return err
}
