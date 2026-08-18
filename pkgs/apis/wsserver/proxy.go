package wsserver

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

type HttpProxy struct {
	ctx           context.Context
	conn          *websocket.Conn
	reqActionType ActionType
	sync.Mutex
	waiting map[string]*proxyWaiter
}

var _ http.RoundTripper = (*HttpProxy)(nil)

func NewProxy(conn *websocket.Conn, reqActionType ActionType) *HttpProxy {
	return &HttpProxy{
		conn:          conn,
		reqActionType: reqActionType,
		waiting:       make(map[string]*proxyWaiter),
	}
}

func (hp *HttpProxy) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	preq, err := NewProxyRequest(req)
	if err != nil {
		logs.ErrorContextf(hp.ctx, "[HttpProxy] new proxy request failed, %s", err)
		return
	}

	wait := hp.Add(preq.TXID)
	defer hp.Remove(preq.TXID)

	err = SendMessage(hp.conn, hp.reqActionType, preq)
	if err != nil {
		logs.ErrorContextf(hp.ctx, "[HttpProxy] send proxy request failed, %s", err)
		return
	}

	return wait.Wait(hp.ctx)
}

func (hp *HttpProxy) Add(txid string) *proxyWaiter {
	waiter := &proxyWaiter{
		txid: txid,
		done: make(chan *ProxyResponse),
	}

	hp.Lock()
	defer hp.Unlock()
	hp.waiting[txid] = waiter
	return waiter
}

func (hp *HttpProxy) Remove(txid string) {
	hp.Lock()
	defer hp.Unlock()
	delete(hp.waiting, txid)
}

type proxyWaiter struct {
	txid string
	done chan *ProxyResponse
}

func (pw *proxyWaiter) Close() {
	select {
	case <-pw.done:
	default:
		close(pw.done)
	}
}

func (pw *proxyWaiter) Wait(ctx context.Context) (resp *http.Response, err error) {
	defer pw.Close()
	tc := time.NewTimer(time.Second * 10)
	defer tc.Stop()
	select {
	case <-tc.C:
		logs.ErrorContextf(ctx, "[HttpProxy] timeout, txid: %s", pw.txid)
		return nil, http.ErrHandlerTimeout
	case <-lifecycle.Std().C():
		logs.ErrorContextf(ctx, "[HttpProxy] lifecycle exit, txid: %s", pw.txid)
		return nil, http.ErrServerClosed
	case presp := <-pw.done:
		return presp.ToHTTPResponse()
	}
}

func (hp *HttpProxy) ProxyResponseHandler(presp *ProxyResponse) {
	logs.InfoContextf(hp.ctx, "[ProxyResponseHandler] got responese txid: %s, status %v", presp.TXID, presp.Status)
	txid := presp.TXID
	waiter, ok := hp.waiting[txid]
	if !ok {
		logs.WarnContextf(hp.ctx, "[ProxyResponseHandler] unknown txid: %s", txid)
		return
	}
	if waiter.done == nil {
		logs.WarnContextf(hp.ctx, "[ProxyResponseHandler] waiter.done is nil")
		return
	}
	waiter.done <- presp
}
