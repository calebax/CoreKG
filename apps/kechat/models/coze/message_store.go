package coze

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ygpkg/yg-go/dbtools/redispool"
)

type storedMessage struct {
	Req CreateMessageRequest `json:"req"`
}

const (
	cozeMessageStoreRedisKeyPrefix = "kechat:coze:message:"
	cozeMessageStoreTTL            = 30 * time.Minute
)

func cozeMessageStoreKey(messageID string) string {
	return cozeMessageStoreRedisKeyPrefix + messageID
}

func storeMessage(req CreateMessageRequest) (string, error) {
	messageID := uuid.NewString()
	msg := &storedMessage{
		Req: req,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	if err := redispool.Redis().Set(context.Background(), cozeMessageStoreKey(messageID), data, cozeMessageStoreTTL).Err(); err != nil {
		return "", err
	}
	return messageID, nil
}

func getMessage(messageID string) (*storedMessage, bool, error) {
	data, err := redispool.Redis().Get(context.Background(), cozeMessageStoreKey(messageID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var msg storedMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, false, err
	}
	return &msg, true, nil
}

func deleteMessage(messageID string) error {
	if messageID == "" {
		return nil
	}
	return redispool.Redis().Del(context.Background(), cozeMessageStoreKey(messageID)).Err()
}
