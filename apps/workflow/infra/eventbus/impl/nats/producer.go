/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nats

import (
	"context"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/insmtx/corekg/apps/workflow/infra/eventbus"
	"github.com/ygpkg/yg-go/logs"
	"github.com/insmtx/corekg/apps/workflow/pkg/taskgroup"
)

type producerImpl struct {
	nc           *nats.Conn
	js           nats.JetStreamContext
	useJetStream bool
	topic        string // Store the topic for this producer instance
	closed       bool
	mu           sync.RWMutex
}

// NewProducer creates a new NATS producer
func NewProducer(serverURL, topic, group string) (eventbus.Producer, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("server URL is empty")
	}

	if topic == "" {
		return nil, fmt.Errorf("topic is empty")
	}

	// Set up NATS connection options
	opts := []nats.Option{
		nats.Name("coze-studio-producer"),
		nats.MaxReconnects(-1), // Unlimited reconnects
	}

	appCfg := conf.GetAppConfig()
	natsConf := appCfg.Workflow.MQ.NATS

	if natsConf.JWTToken != "" {
		opts = append(opts, nats.UserJWTAndSeed(natsConf.JWTToken, natsConf.NKeySeed))
	} else if natsConf.Username != "" {
		opts = append(opts, nats.UserInfo(natsConf.Username, natsConf.Password))
	} else if natsConf.Token != "" {
		opts = append(opts, nats.Token(natsConf.Token))
	}

	nc, err := nats.Connect(serverURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS failed: %w", err)
	}

	useJetStream := natsConf.UseJetStream

	producer := &producerImpl{
		nc:           nc,
		useJetStream: useJetStream,
		topic:        topic, // Store the topic for this producer instance
		closed:       false,
	}

	// Initialize JetStream if needed
	if useJetStream {
		js, err := nc.JetStream()
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("create JetStream context failed: %w", err)
		}
		producer.js = js
	}

	return producer, nil
}

// Send sends a single message using the stored topic
func (p *producerImpl) Send(ctx context.Context, body []byte, opts ...eventbus.SendOpt) error {
	return p.BatchSend(ctx, [][]byte{body}, opts...)
}

// BatchSend sends multiple messages using the stored topic
func (p *producerImpl) BatchSend(ctx context.Context, bodyArr [][]byte, opts ...eventbus.SendOpt) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	if len(bodyArr) == 0 {
		return fmt.Errorf("no messages to send")
	}

	// Use the stored topic
	topic := p.topic
	if topic == "" {
		return fmt.Errorf("topic is not set")
	}

	// Parse producer options
	option := &eventbus.SendOption{}
	for _, opt := range opts {
		opt(option)
	}

	if p.useJetStream {
		return p.batchSendJetStream(ctx, topic, bodyArr, option)
	} else {
		return p.batchSendCore(ctx, topic, bodyArr, option)
	}
}

// batchSendJetStream sends messages using JetStream for persistence
func (p *producerImpl) batchSendJetStream(ctx context.Context, topic string, messages [][]byte, option *eventbus.SendOption) error {
	// Ensure Stream exists
	if err := ensureStream(p.js, topic); err != nil {
		return fmt.Errorf("ensure stream failed: %w", err)
	}

	// Use TaskGroup to wait for all async publishes
	tg := taskgroup.NewTaskGroup(ctx, min(len(messages), 5))

	for i, message := range messages {
		tg.Go(func() error {
			// Prepare publish options
			pubOpts := []nats.PubOpt{}

			// Add message ID for deduplication if sharding key is provided
			if option.ShardingKey != nil && *option.ShardingKey != "" {
				msgID := fmt.Sprintf("%s-%d", *option.ShardingKey, i)
				pubOpts = append(pubOpts, nats.MsgId(msgID))
			}

			// Add context for timeout
			pubOpts = append(pubOpts, nats.Context(ctx))

			// Publish message asynchronously
			_, err := p.js.Publish(topic, message, pubOpts...)
			if err != nil {
				return fmt.Errorf("publish message %d failed: %w", i, err)
			}
			return nil
		})
	}

	// Wait for all messages to be sent
	if err := tg.Wait(); err != nil {
		return err
	}

	logs.DebugContextf(ctx,"successfully sent %d messages to NATS JetStream topic: %s", len(messages), topic)
	return nil
}

// batchSendCore sends messages using core NATS for simple pub/sub
func (p *producerImpl) batchSendCore(ctx context.Context, topic string, messages [][]byte, option *eventbus.SendOption) error {
	// Use TaskGroup to wait for all async publishes
	tg := taskgroup.NewTaskGroup(ctx, min(len(messages), 5))

	for i, message := range messages {
		tg.Go(func() error {
			// For core NATS, we can add headers if sharding key is provided
			if option.ShardingKey != nil && *option.ShardingKey != "" {
				// Create message with headers
				natsMsg := &nats.Msg{
					Subject: topic,
					Data:    message,
					Header:  nats.Header{},
				}
				natsMsg.Header.Set("Sharding-Key", *option.ShardingKey)

				err := p.nc.PublishMsg(natsMsg)
				if err != nil {
					return fmt.Errorf("publish message %d with header failed: %w", i, err)
				}
			} else {
				// Simple publish without headers
				err := p.nc.Publish(topic, message)
				if err != nil {
					return fmt.Errorf("publish message %d failed: %w", i, err)
				}
			}
			return nil
		})
	}

	// Wait for all messages to be sent
	if err := tg.Wait(); err != nil {
		return err
	}

	// Flush to ensure all messages are sent
	if err := p.nc.Flush(); err != nil {
		return fmt.Errorf("flush NATS connection failed: %w", err)
	}
	logs.DebugContextf(ctx,"successfully sent %d messages to NATS core topic: %s", len(messages), topic)
	return nil
}

// Close closes the producer and releases resources
func (p *producerImpl) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	if p.nc != nil {
		// Drain connection to ensure all pending messages are sent
		if err := p.nc.Drain(); err != nil {
			logs.Warnf("drain NATS connection failed: %v", err)
		}
		p.nc.Close()
	}

	logs.Infof("NATS producer closed successfully")
	return nil
}
