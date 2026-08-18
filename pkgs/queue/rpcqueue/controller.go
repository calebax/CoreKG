package rpcqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/logs"
)

// RPCQueue is the class for all the RPCs
type RPCQueue struct {
	ctx              context.Context
	ctxCancel        context.CancelFunc
	connUrl          string
	conn             *amqp.Connection
	rabbitChan       *amqp.Channel
	exchangeName     string
	requestQueueName string
	replyQueue       amqp.Queue
	responseChLocker sync.RWMutex
	responseCh       map[string]chan string
	timeout          time.Duration
}

// NewRPCQueue creates a new RPC Queue todo copy context
func NewRPCQueue(ctx context.Context, connUrl string, exchangeName, reqQueueName string, timeout time.Duration) (rq *RPCQueue, err error) {
	rq = &RPCQueue{
		ctx:              ctx,
		connUrl:          connUrl,
		exchangeName:     exchangeName,
		requestQueueName: reqQueueName,
		responseCh:       make(map[string]chan string),
		timeout:          timeout,
	}
	// 连接RabbitMQ
	err = rq.Init()
	if err != nil {
		logs.ErrorContextf(ctx, "rabbitMq init err: %v", err)
		return nil, err
	}

	return rq, err
}

func (rq *RPCQueue) Init() (err error) {
	logs.InfoContextf(rq.ctx, "--------init,queue = %s, conn = %s", rq.requestQueueName, rq.connUrl)
	ctx, cancel := context.WithCancel(context.WithoutCancel(rq.ctx))
	rq.ctx = ctx
	rq.ctxCancel = cancel

	rq.conn, err = amqp091.Dial(rq.connUrl)
	if err != nil {
		logs.ErrorContextf(rq.ctx, "rabbitMq conn err: %v", err)
		return err
	}
	rq.rabbitChan, err = rq.conn.Channel()
	if err != nil {
		logs.ErrorContextf(rq.ctx, "rabbitMq conn err: %v", err)
		return err
	}
	// 声明请求消息队列
	_, err = rq.rabbitChan.QueueDeclare(
		rq.requestQueueName, true, false, false, false, nil,
	)
	if err != nil {
		logs.ErrorContextf(rq.ctx, "rabbitMq req queue declare failed: %v", err)
		return err
	}
	// 创建临时响应队列（自动删除）
	rq.replyQueue, err = rq.rabbitChan.QueueDeclare(
		"", false, true, true, false, nil,
	)
	if err != nil {
		logs.ErrorContextf(rq.ctx, "rabbitMq resp queue declare failed: %v", err)
		return err
	}
	// 消费回复队列
	msgs, err := rq.rabbitChan.ConsumeWithContext(rq.ctx,
		rq.replyQueue.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		logs.ErrorContextf(rq.ctx, "rabbitMq consume reply queue failed: %v", err)
		return err
	}
	go rq.ConsumeReplyRoutine(msgs)
	rq.WatchConn()
	logs.InfoContextf(rq.ctx, "--------init success")
	return nil
}

func (rq *RPCQueue) WatchConn() {
	// 监听连接状态
	connErrChan := make(chan *amqp.Error, 1)
	rq.conn.NotifyClose(connErrChan)
	go func() {
		for {
			select {
			case err := <-connErrChan:
				if err != nil {
					logs.ErrorContextf(rq.ctx, "rabbitMQ connection closed : %v", err)
					rq.ReopenConn()
					return
				}
			case <-rq.ctx.Done():
				rq.Release()
				logs.InfoContextf(rq.ctx, "rabbitMQ context.Done(), exit the watch routine")
				return
			}
		}
	}()
}

func (rq *RPCQueue) Release() {
	rq.ctxCancel()
	rq.rabbitChan.Close()
	rq.conn.Close()
}

// SendRequest 发送请求，并且监听等待响应。请求和响应都是JSON序列化的对象。
func (rq *RPCQueue) SendRequest(corrID string, queueNamereqBody interface{}) (string, error) {
	if corrID == "" {
		corrID = encryptor.GenerateUUID()
	}
	ch := make(chan string)
	rq.responseChLocker.Lock()
	rq.responseCh[corrID] = ch
	rq.responseChLocker.Unlock()
	defer func() {
		rq.responseChLocker.Lock()
		delete(rq.responseCh, corrID)
		rq.responseChLocker.Unlock()
		close(ch)
	}()

	{
		request, err := json.Marshal(queueNamereqBody)
		if err != nil {
			logs.ErrorContextf(rq.ctx, "marshal request err: %v", err)
			return "", err
		}
		err = rq.rabbitChan.Publish(
			rq.exchangeName, rq.requestQueueName, false, false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrID,
				ReplyTo:       rq.replyQueue.Name,
				Body:          []byte(request),
			},
		)
		if err != nil {
			logs.ErrorContextf(rq.ctx, "rabbitMq publish failed: %v", err)
			return "", err
		}
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(rq.timeout):
		return "", fmt.Errorf("request queue(%s) timeout", rq.requestQueueName)
	}
}

// ConsumeReplyRoutine 消费响应
func (rq *RPCQueue) ConsumeReplyRoutine(msgs <-chan amqp.Delivery) {
	for {
		select {
		case msg := <-msgs:
			corrID := msg.CorrelationId
			rq.responseChLocker.RLock()
			ch, ok := rq.responseCh[corrID]
			rq.responseChLocker.RUnlock()
			if ok {
				select {
				case ch <- string(msg.Body):
					// 成功写入
				default:
					logs.WarnContextf(rq.ctx, "receive message but send failed(timeout or something), msg=%s", msg.Body)
				}
			} else {
				logs.WarnContextf(rq.ctx, "receive message but no channel found, corrID=%s, msg=%s", corrID, msg.Body)
			}
		case <-rq.ctx.Done():
			logs.InfoContextf(rq.ctx, "rabbitMQ context.Done(), exit the consume routine")
			return
		}
	}
}

// ReopenConn 重新打开连接
func (rq *RPCQueue) ReopenConn() {
	logs.InfoContextf(rq.ctx, "-------reopen connection")
	rq.Release()
	if err := rq.Init(); err != nil {
		logs.ErrorContextf(rq.ctx, " rabbitMq reopen connection failed: %v", err)
		time.Sleep(5 * time.Second)
		rq.ReopenConn()
	}
	logs.InfoContextf(rq.ctx, "-------reopen connection success")
}
