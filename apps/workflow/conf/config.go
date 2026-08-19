package conf

import (
	ygconfig "github.com/ygpkg/yg-go/config"
)

type AppConfig struct {
	ygconfig.CoreConfig `yaml:",inline"`
	Workflow            WorkflowConfig `yaml:"workflow"`
}

type WorkflowConfig struct {
	Enabled            bool             `yaml:"enabled"` // 聚合模式下是否拉启 workflow；false 时不启动
	Required           bool             `yaml:"required"` // 启动失败时是否拖垮宿主进程
	HttpAddr           string           `yaml:"http_addr"` // 独立监听地址（聚合模式优先于此字段）
	LogLevel           string           `yaml:"log_level"`
	MaxRequestBodySize int64            `yaml:"max_request_body_size"`
	ServerHost         string           `yaml:"server_host"`
	AdminUins          string           `yaml:"admin_uins"`
	SSL                SSLConfig        `yaml:"ssl"`
	Redis              RedisConfig      `yaml:"redis"`
	Elasticsearch      ESConfig         `yaml:"elasticsearch"`
	Storage            StorageConfig    `yaml:"storage"`
	MQ                 MQConfig         `yaml:"mq"`
	Upload             UploadConfig     `yaml:"upload"`
}

type SSLConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type ESConfig struct {
	Addr              string `yaml:"addr"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	Version           string `yaml:"version"`
	NumberOfShards    string `yaml:"number_of_shards"`
	NumberOfReplicas  string `yaml:"number_of_replicas"`
}

type StorageConfig struct {
	Type               string       `yaml:"type"`
	UploadHTTPScheme   string       `yaml:"upload_http_scheme"`
	Bucket             string       `yaml:"bucket"`
	MinIO              MinIOConfig  `yaml:"minio"`
	TOS                TOSConfig    `yaml:"tos"`
	S3                 S3Config     `yaml:"s3"`
}

type MinIOConfig struct {
	AK       string `yaml:"ak"`
	SK       string `yaml:"sk"`
	Endpoint string `yaml:"endpoint"`
	Region   string `yaml:"region"`
	APIHost  string `yaml:"api_host"`
}

type TOSConfig struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Endpoint  string `yaml:"endpoint"`
	Region    string `yaml:"region"`
}

type S3Config struct {
	AccessKey      string `yaml:"access_key"`
	SecretKey      string `yaml:"secret_key"`
	Endpoint       string `yaml:"endpoint"`
	Region         string `yaml:"region"`
	BucketEndpoint string `yaml:"bucket_endpoint"`
}

type MQConfig struct {
	Type       string         `yaml:"type"`
	NameServer string         `yaml:"name_server"`
	RMQ        RMQConfig      `yaml:"rmq"`
	Pulsar     PulsarConfig   `yaml:"pulsar"`
	NATS       NATSConfig     `yaml:"nats"`
}

type RMQConfig struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type PulsarConfig struct {
	ServiceURL string `yaml:"service_url"`
	JWTToken   string `yaml:"jwt_token"`
}

type NATSConfig struct {
	JWTToken    string `yaml:"jwt_token"`
	NKeySeed    string `yaml:"nkey_seed"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	Token       string `yaml:"token"`
	UseJetStream bool   `yaml:"use_jetstream"`
}

type UploadConfig struct {
	ComponentType string        `yaml:"component_type"`
	ImageX        ImageXConfig  `yaml:"imagex"`
}

type ImageXConfig struct {
	AK         string `yaml:"ak"`
	SK         string `yaml:"sk"`
	ServerID   string `yaml:"server_id"`
	Domain     string `yaml:"domain"`
	Template   string `yaml:"template"`
	UploadHost string `yaml:"upload_host"`
}

var stdAppConfig *AppConfig

func SetAppConfig(cfg *AppConfig) {
	stdAppConfig = cfg
}

func GetAppConfig() *AppConfig {
	if stdAppConfig == nil {
		return &AppConfig{}
	}
	return stdAppConfig
}
