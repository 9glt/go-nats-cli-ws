package natscliws

import (
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	nats "github.com/nats-io/nats.go"
)

// New connects to nats via websocket dialer
func New(ws string, n string) (*MQ, error) {
	opts := []nats.Option{
		nats.SetCustomDialer(&customDialer{ws}),
		nats.ReconnectWait(1 * time.Second),
		nats.DontRandomize(),
		nats.MaxReconnects(1<<31 - 1),
	}
	nc, err := nats.Connect(n, opts...)
	return &MQ{nc}, err
}

type customDialer struct {
	url string
}

func (cd *customDialer) Dial(network, address string) (net.Conn, error) {

	u, _ := url.Parse(cd.url)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.UnderlyingConn(), nil
}

type MQ struct {
	conn *nats.Conn
}

func (mq *MQ) UnderlyingConn() *nats.Conn {
	return mq.conn
}

func (mq *MQ) Subscribe(t string, cb MsgHandler) (*Subscription, error) {
	sub, err := mq.conn.Subscribe(t, func(msg *nats.Msg) {
		cb(&Msg{*msg})
	})
	if err != nil {
		return nil, err
	}

	return &Subscription{*sub}, nil
}

func (mq *MQ) QueueSubscribe(t, q string, cb MsgHandler) (*Subscription, error) {
	sub, err := mq.conn.QueueSubscribe(t, q, func(msg *nats.Msg) {
		cb(&Msg{*msg})
	})
	if err != nil {
		return nil, err
	}
	return &Subscription{*sub}, nil
}

func (mq *MQ) Publish(t string, pl []byte) error {
	return mq.conn.Publish(t, pl)
}

func (mq *MQ) Flush() error {
	return mq.conn.Flush()
}

func (mq *MQ) FlushTimeout(d time.Duration) error {
	return mq.conn.FlushTimeout(d)
}

type MsgHandler func(msg *Msg)

type Msg struct {
	nats.Msg
}

type Subscription struct {
	nats.Subscription
}
