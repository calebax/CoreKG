package models

import (
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

type Memory struct {
	mu sync.RWMutex

	Messages         []*Message
	CallbackExecInfo map[string]*CallbackExecution
	TempMsgs         map[string]*TempMessage
}

type Message struct {
	MessageId   string          `json:"message_id"`
	MessageType string          `json:"message_type"`
	MessageTime int64           `json:"message_time"`
	Payload     *schema.Message `json:"payload"`
}

func NewMemory() *Memory {
	return &Memory{
		Messages:         make([]*Message, 0),
		CallbackExecInfo: make(map[string]*CallbackExecution),
		TempMsgs:         make(map[string]*TempMessage),
	}
}

// 临时暂存消息
type TempMessage struct {
	MessageId   string            `json:"message_id"`
	MessageType string            `json:"message_type"`
	Payloads    []*schema.Message `json:"payloads"`
}

type CallbackExecution struct {
	MessageId   string `json:"messageId"`
	MessageType string `json:"messageType"`
	Name        string `json:"name"`
}

func (m *Memory) CreateMessageId(msgType string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	msgId := uuid.NewString()
	m.TempMsgs[msgId] = &TempMessage{
		MessageId:   msgId,
		MessageType: msgType,
	}
	return msgId
}

func (m *Memory) AddMessageWithType(msgType string, msg *schema.Message) {
	m.AddMessage("", msgType, msg)
}

func (m *Memory) AddMessage(msgId string, msgType string, msg *schema.Message) {
	if msgId == "" {
		msgId = uuid.NewString()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, &Message{
		MessageId:   msgId,
		MessageType: msgType,
		MessageTime: time.Now().Unix(),
		Payload:     msg,
	})
}

func (m *Memory) AppendTempPayload(messageID string, msg *schema.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.TempMsgs[messageID] != nil {
		m.TempMsgs[messageID].Payloads = append(m.TempMsgs[messageID].Payloads, msg)
	}
}

func (m *Memory) PrivateAddMessage(msg *Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Messages = append(m.Messages, msg)
}

func (m *Memory) GetLastMessage() *Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Messages) == 0 {
		return nil
	}
	return m.Messages[len(m.Messages)-1]
}

func (m *Memory) GetLlmMessages() []*schema.Message {
	var res []*schema.Message
	for _, msg := range m.Messages {
		res = append(res, msg.Payload)
	}
	return res
}

func (m *Memory) FlushMsg(messageID string) *Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	tempMsg := m.TempMsgs[messageID]
	if tempMsg == nil {
		return nil
	}

	defer func() {
		delete(m.TempMsgs, messageID)
	}()

	if len(tempMsg.Payloads) == 0 {
		return nil
	}

	var concatedMsg *schema.Message
	if len(tempMsg.Payloads) == 1 {
		concatedMsg = tempMsg.Payloads[0]
	} else {
		concatedMsg, _ = schema.ConcatMessages(tempMsg.Payloads)
	}

	if concatedMsg == nil {
		return nil
	}

	msg := &Message{
		MessageId:   messageID,
		MessageType: tempMsg.MessageType,
		MessageTime: time.Now().Unix(),
		Payload:     concatedMsg,
	}

	m.Messages = append(m.Messages, msg)
	return msg
}

func (m *Memory) Clear() {
	m.Messages = nil
}
