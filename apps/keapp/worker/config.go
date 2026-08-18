package worker

import "time"

type Config struct {
	Concurrency       int           `json:"concurrency" yaml:"concurrency"`
	CancelCheckInterval int         `json:"cancel_check_interval" yaml:"cancelCheckInterval"`
	StreamName        string        `json:"stream_name" yaml:"streamName"`
	Subject           string        `json:"subject" yaml:"subject"`
	ConsumerName      string        `json:"consumer_name" yaml:"consumerName"`
	MaxDeliver        int           `json:"max_deliver" yaml:"maxDeliver"`
	AckWait           time.Duration `json:"ack_wait" yaml:"ackWait"`
}

func DefaultConfig() Config {
	return Config{
		Concurrency:       3,
		CancelCheckInterval: 10,
		StreamName:        "KEAPP_CRAWL_STREAM",
		Subject:           "keapp.crawl.trigger",
		ConsumerName:      "keapp-crawl-worker",
		MaxDeliver:        3,
		AckWait:           5 * time.Minute,
	}
}
