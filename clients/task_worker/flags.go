package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/spf13/cobra"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/dbtools/estool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/storage"
)

var (
	s3AccessKeyID     string
	s3SecretAccessKey string
	s3EndpointURL     string
	s3BucketName      string
)

func withS3Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s3AccessKeyID, "s3_access_key_id", os.Getenv("S3_ACCESS_KEY_ID"), "S3 access key ID")
	cmd.Flags().StringVar(&s3SecretAccessKey, "s3_secret_access_key", os.Getenv("S3_SECRET_ACCESS_KEY"), "S3 secret access key")
	cmd.Flags().StringVar(&s3EndpointURL, "s3_endpoint_url", os.Getenv("S3_ENDPOINT_URL"), "S3 endpoint URL")
	cmd.Flags().StringVar(&s3BucketName, "s3_bucket_name", os.Getenv("S3_BUCKET_NAME"), "S3 bucket name")
}

func loadS3Cli() (*storage.S3Fs, error) {
	cfg := config.S3StorageConfig{
		AccessKeyID:     s3AccessKeyID,
		SecretAccessKey: s3SecretAccessKey,
		EndPoint:        s3EndpointURL,
		Bucket:          s3BucketName,
		Region:          "us-east-1", // 默认区域，可根据需要修改
		UsePathStyle:    true,        // 是否使用路径风格的URL
	}
	s3Cli, err := storage.NewS3Fs(cfg, config.StorageOption{})
	if err != nil {
		return nil, err
	}
	return s3Cli, nil
}

var (
	esAddresses string
	esUsername  string
	esPassword  string
)

func withESFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&esAddresses, "es_host", os.Getenv("ES_HOST"), "Elasticsearch host")
	cmd.Flags().StringVar(&esUsername, "es_username", os.Getenv("ES_USERNAME"), "Elasticsearch username")
	cmd.Flags().StringVar(&esPassword, "es_password", os.Getenv("ES_PASSWORD"), "Elasticsearch password")
}

func loadESClient() (*elasticsearch.Client, error) {
	cfg := config.ESConfig{
		Addresses:  []string{esAddresses},
		Username:   esUsername,
		Password:   esPassword,
		MaxRetries: 3, // 最大重试次数
	}
	cli, err := estool.InitES(cfg)
	if err != nil {
		panic(err)
	}
	return cli, nil
}

var (
	model                                string
	mindMapMD, abstractMD, descriptionMD string
)

func withAgentFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&model, "model", os.Getenv("MODEL_NAME"), "name about agent model")
	cmd.Flags().StringVar(&mindMapMD, "mindmap", "", "name of mindmap agent")
	cmd.Flags().StringVar(&abstractMD, "abstract", "", "name of abstract agent")
	cmd.Flags().StringVar(&descriptionMD, "description", "", "name of description agent")
}

var (
	//bucketName:bucketCli
	s3m map[string]*storage.S3Fs = nil

	// esClient 用于存储 Elasticsearch 客户端
	esClient *elasticsearch.Client

	//once
	once sync.Once

	//App config
	appCfg = &AppConfig{}
)

func loadConfig(ctx context.Context, configPath string) {
	once.Do(func() {
		if len(configPath) == 0 {
			panic("config file path is empty")
		}
		appCfg = &AppConfig{}
		if err := config.LoadYamlLocalFile(configPath, appCfg); err != nil {
			panic(fmt.Errorf("failed to load yaml file: %w", err))
		}

		NewS3Cli(appCfg.S3Storages)

		NewEsCli(appCfg.Elasticsearch)

		logs.DebugContextf(ctx, "Configuration loading completed.\n %v", appCfg)
	})
}

// NewNBCli will new a nebula cli, if var appCfg is nil, this func would panic
func NewNBCli(cfg *nbgraph.NebulaConf) *nbgraph.NebulaCli {
	if cfg == nil {
		panic("NebulaConf config is nil")
	}
	hostAddress := nebula.HostAddress{Host: cfg.Address, Port: cfg.Port}
	hostList := []nebula.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebula.GetDefaultConf()
	fmt.Println(testPoolConfig)
	// Initialize connection pool
	pool, err := nebula.NewConnectionPool(hostList, testPoolConfig, nebula.DefaultLogger{})
	if err != nil {
		panic(fmt.Errorf("fail to initialize the connection pool, "+
			"host: %s, port: %d, %s", cfg.Address, cfg.Port, err.Error()))
	}
	// Create session
	session, err := pool.GetSession(cfg.UserName, cfg.Password)
	if err != nil {
		panic(fmt.Errorf("fail to create a new session from connection pool, "+
			"username: %s, password: %s, %s", cfg.UserName, cfg.Password, err.Error()))
	}

	return &nbgraph.NebulaCli{Session: session}
}

// NewS3Cli 根据配置初始化 S3 客户端映射
func NewS3Cli(s3Configs []config.S3StorageConfig) {
	s3m = make(map[string]*storage.S3Fs, len(s3Configs))
	for _, sc := range s3Configs {
		s3Cli, err := storage.NewS3Fs(config.S3StorageConfig{
			EndPoint:        sc.EndPoint,
			AccessKeyID:     sc.AccessKeyID,
			SecretAccessKey: sc.SecretAccessKey,
			Bucket:          sc.Bucket,
			Region:          sc.Region,
			UsePathStyle:    sc.UsePathStyle,
		}, config.StorageOption{})
		if err != nil {
			panic(fmt.Errorf("failed to create S3 client for bucket %s: %w", sc.Bucket, err))
		}
		s3m[sc.Bucket] = s3Cli
	}
	logs.DebugContextf(context.TODO(), "S3 clients loaded for buckets: %+v", s3m)
}

// NewEsCli 根据配置初始化 Elasticsearch 客户端
func NewEsCli(esConfig config.ESConfig) {
	esCfg := elasticsearch.Config{
		Addresses: esConfig.Addresses,
		Username:  esConfig.Username,
		Password:  esConfig.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 忽略 SSL 证书验证
			},
		},
	}
	client, newClientErr := elasticsearch.NewClient(esCfg)
	if newClientErr != nil {
		panic(newClientErr)
	}
	esClient = client
}

type AppConfig struct {
	S3Storages    []config.S3StorageConfig `yaml:"s3_storages"`
	Elasticsearch config.ESConfig          `yaml:"elasticsearch"`
	Nebula        nbgraph.NebulaConf       `yaml:"nebula"`
	Agent         AgentConfig              `yaml:"agent"`
}

type AgentConfig struct {
	APIUrl       string            `yaml:"apiUrl"`
	APIKey       string            `yaml:"apiKey"`
	ChunkSize    uint              `yaml:"chunkSize"`
	MaxWorkers   uint              `yaml:"maxWorkers"`
	MaxTokenSize uint              `yaml:"maxTokenSize"`
	Pool         map[string]string `yaml:"pool"`
}
