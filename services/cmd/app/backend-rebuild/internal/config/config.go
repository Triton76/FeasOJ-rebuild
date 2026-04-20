// Layer: Config (配置层)
// Responsibility: 配置加载(YAML+环境变量)、默认值管理、配置聚合
// Dependency: 不依赖其他层，被 Bootstrap 层调用
package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const defaultAddr = "127.0.0.1:8082"
const defaultJWTIssuer = "feasoj-backend-rebuild"
const defaultJWTExpireHours = "2"
const defaultContestStatusScanSeconds = 30
const defaultConfigPath = "app/backend-rebuild/config.yaml"

type Config struct {
	ConfigPath                 string
	ConfigSource               string
	Addr                      string
	MySQLDSN                  string
	JWTSecret                 string
	JudgeWritebackToken       string
	JWTIssuer                 string
	JWTExpireH                string
	ContestStatusScanSeconds  int
	EnableJudgeWriteback      bool
	EnableScoreboard          bool
	EnableTestcaseAPIs        bool
	EnableRabbitMQQueue       bool
	EnableEmbeddedJudgeWorker bool
	EnableClassWorkflowV2     bool
	EnablePasswordReset       bool
	RabbitMQURL               string
	RabbitMQExchange          string
	RabbitMQMainQueue         string
	RabbitMQRetryQueue        string
	RabbitMQDLQ               string
	RabbitMQWorkerPrefetch    int
}

type fileConfig struct {
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
	JWT struct {
		Secret      string `yaml:"secret"`
		Issuer      string `yaml:"issuer"`
		ExpireHours int    `yaml:"expire_hours"`
	} `yaml:"jwt"`
	Judge struct {
		WritebackToken string `yaml:"writeback_token"`
	} `yaml:"judge"`
	Scheduler struct {
		ContestStatusScanSeconds int `yaml:"contest_status_scan_seconds"`
	} `yaml:"scheduler"`
	FeatureFlags struct {
		EnableJudgeWriteback      *bool `yaml:"enable_judge_writeback"`
		EnableScoreboard          *bool `yaml:"enable_scoreboard"`
		EnableTestcaseAPIs        *bool `yaml:"enable_testcase_apis"`
		EnableRabbitMQQueue       *bool `yaml:"enable_rabbitmq_queue"`
		EnableEmbeddedJudgeWorker *bool `yaml:"enable_embedded_judge_worker"`
		EnableClassWorkflowV2     *bool `yaml:"enable_class_workflow_v2"`
		EnablePasswordReset       *bool `yaml:"enable_password_reset"`
	} `yaml:"feature_flags"`
	RabbitMQ struct {
		URL        string `yaml:"url"`
		Exchange   string `yaml:"exchange"`
		MainQueue  string `yaml:"main_queue"`
		RetryQueue string `yaml:"retry_queue"`
		DLQ        string `yaml:"dlq"`
		Prefetch   int    `yaml:"worker_prefetch"`
	} `yaml:"rabbitmq"`
}

func Load() (Config, error) {
	cfg := Config{
		ConfigSource:              "defaults",
		Addr:                      defaultAddr,
		JWTIssuer:                 defaultJWTIssuer,
		JWTExpireH:                defaultJWTExpireHours,
		ContestStatusScanSeconds:  defaultContestStatusScanSeconds,
		EnableJudgeWriteback:      true,
		EnableScoreboard:          true,
		EnableTestcaseAPIs:        false,
		EnableRabbitMQQueue:       false,
		EnableEmbeddedJudgeWorker: true,
		EnableClassWorkflowV2:     true,
		EnablePasswordReset:       false,
		RabbitMQExchange:          "judge.submission.exchange",
		RabbitMQMainQueue:         "judge.submission.main",
		RabbitMQRetryQueue:        "judge.submission.retry",
		RabbitMQDLQ:               "judge.submission.dlq",
		RabbitMQWorkerPrefetch:    1,
	}

	if err := loadFromYAML(&cfg); err != nil {
		return Config{}, err
	}
	loadFromEnv(&cfg)
	if err := validateStrict(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadFromYAML(cfg *Config) error {
	path := os.Getenv("BACKEND_REBUILD_CONFIG")
	if path == "" {
		path = defaultConfigPath
	}
	cfg.ConfigPath = path

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.ConfigSource = "defaults"
			return nil
		}
		return fmt.Errorf("read config yaml failed: %w", err)
	}

	var fc fileConfig
	if err := yaml.Unmarshal(content, &fc); err != nil {
		return fmt.Errorf("parse config yaml failed: %w", err)
	}
	cfg.ConfigSource = "yaml:" + path

	if fc.Server.Addr != "" {
		cfg.Addr = fc.Server.Addr
	}
	if fc.MySQL.DSN != "" {
		cfg.MySQLDSN = fc.MySQL.DSN
	}
	if fc.JWT.Secret != "" {
		cfg.JWTSecret = fc.JWT.Secret
	}
	if fc.Judge.WritebackToken != "" {
		cfg.JudgeWritebackToken = fc.Judge.WritebackToken
	}
	if fc.JWT.Issuer != "" {
		cfg.JWTIssuer = fc.JWT.Issuer
	}
	if fc.JWT.ExpireHours > 0 {
		cfg.JWTExpireH = strconv.Itoa(fc.JWT.ExpireHours)
	}
	if fc.Scheduler.ContestStatusScanSeconds > 0 {
		cfg.ContestStatusScanSeconds = fc.Scheduler.ContestStatusScanSeconds
	}
	if fc.FeatureFlags.EnableJudgeWriteback != nil {
		cfg.EnableJudgeWriteback = *fc.FeatureFlags.EnableJudgeWriteback
	}
	if fc.FeatureFlags.EnableScoreboard != nil {
		cfg.EnableScoreboard = *fc.FeatureFlags.EnableScoreboard
	}
	if fc.FeatureFlags.EnableTestcaseAPIs != nil {
		cfg.EnableTestcaseAPIs = *fc.FeatureFlags.EnableTestcaseAPIs
	}
	if fc.FeatureFlags.EnableRabbitMQQueue != nil {
		cfg.EnableRabbitMQQueue = *fc.FeatureFlags.EnableRabbitMQQueue
	}
	if fc.FeatureFlags.EnableEmbeddedJudgeWorker != nil {
		cfg.EnableEmbeddedJudgeWorker = *fc.FeatureFlags.EnableEmbeddedJudgeWorker
	}
	if fc.FeatureFlags.EnableClassWorkflowV2 != nil {
		cfg.EnableClassWorkflowV2 = *fc.FeatureFlags.EnableClassWorkflowV2
	}
	if fc.FeatureFlags.EnablePasswordReset != nil {
		cfg.EnablePasswordReset = *fc.FeatureFlags.EnablePasswordReset
	}
	if fc.RabbitMQ.URL != "" {
		cfg.RabbitMQURL = fc.RabbitMQ.URL
	}
	if fc.RabbitMQ.Exchange != "" {
		cfg.RabbitMQExchange = fc.RabbitMQ.Exchange
	}
	if fc.RabbitMQ.MainQueue != "" {
		cfg.RabbitMQMainQueue = fc.RabbitMQ.MainQueue
	}
	if fc.RabbitMQ.RetryQueue != "" {
		cfg.RabbitMQRetryQueue = fc.RabbitMQ.RetryQueue
	}
	if fc.RabbitMQ.DLQ != "" {
		cfg.RabbitMQDLQ = fc.RabbitMQ.DLQ
	}
	if fc.RabbitMQ.Prefetch > 0 {
		cfg.RabbitMQWorkerPrefetch = fc.RabbitMQ.Prefetch
	}

	return nil
}

func loadFromEnv(cfg *Config) {
	envOverride := false
	if v := os.Getenv("BACKEND_REBUILD_ADDR"); v != "" {
		cfg.Addr = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_MYSQL_DSN"); v != "" {
		cfg.MySQLDSN = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_JUDGE_WRITEBACK_TOKEN"); v != "" {
		cfg.JudgeWritebackToken = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_JWT_ISSUER"); v != "" {
		cfg.JWTIssuer = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_JWT_EXPIRE_HOURS"); v != "" {
		cfg.JWTExpireH = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_CONTEST_STATUS_SCAN_SECONDS"); v != "" {
		envOverride = true
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			cfg.ContestStatusScanSeconds = sec
		}
	}
	if v := os.Getenv("BACKEND_REBUILD_ENABLE_JUDGE_WRITEBACK"); v != "" {
		cfg.EnableJudgeWriteback = v == "1" || v == "true" || v == "TRUE"
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_ENABLE_SCOREBOARD"); v != "" {
		cfg.EnableScoreboard = v == "1" || v == "true" || v == "TRUE"
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_ENABLE_TESTCASE_APIS"); v != "" {
		cfg.EnableTestcaseAPIs = v == "1" || v == "true" || v == "TRUE"
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_ENABLE_RABBITMQ_QUEUE"); v != "" {
		cfg.EnableRabbitMQQueue = v == "1" || v == "true" || v == "TRUE"
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_ENABLE_EMBEDDED_JUDGE_WORKER"); v != "" {
		cfg.EnableEmbeddedJudgeWorker = v == "1" || v == "true" || v == "TRUE"
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_ENABLE_CLASS_WORKFLOW_V2"); v != "" {
		cfg.EnableClassWorkflowV2 = v == "1" || v == "true" || v == "TRUE"
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_ENABLE_PASSWORD_RESET"); v != "" {
		cfg.EnablePasswordReset = v == "1" || v == "true" || v == "TRUE"
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_RABBITMQ_URL"); v != "" {
		cfg.RabbitMQURL = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_RABBITMQ_EXCHANGE"); v != "" {
		cfg.RabbitMQExchange = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_RABBITMQ_MAIN_QUEUE"); v != "" {
		cfg.RabbitMQMainQueue = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_RABBITMQ_RETRY_QUEUE"); v != "" {
		cfg.RabbitMQRetryQueue = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_RABBITMQ_DLQ"); v != "" {
		cfg.RabbitMQDLQ = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_RABBITMQ_WORKER_PREFETCH"); v != "" {
		envOverride = true
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RabbitMQWorkerPrefetch = n
		}
	}

	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = defaultJWTIssuer
	}
	if cfg.JWTExpireH == "" {
		cfg.JWTExpireH = defaultJWTExpireHours
	}
	if cfg.ContestStatusScanSeconds <= 0 {
		cfg.ContestStatusScanSeconds = defaultContestStatusScanSeconds
	}
	if cfg.RabbitMQExchange == "" {
		cfg.RabbitMQExchange = "judge.submission.exchange"
	}
	if cfg.RabbitMQMainQueue == "" {
		cfg.RabbitMQMainQueue = "judge.submission.main"
	}
	if cfg.RabbitMQRetryQueue == "" {
		cfg.RabbitMQRetryQueue = "judge.submission.retry"
	}
	if cfg.RabbitMQDLQ == "" {
		cfg.RabbitMQDLQ = "judge.submission.dlq"
	}
	if cfg.RabbitMQWorkerPrefetch <= 0 {
		cfg.RabbitMQWorkerPrefetch = 1
	}
	if envOverride {
		if cfg.ConfigSource == "" {
			cfg.ConfigSource = "env"
		} else {
			cfg.ConfigSource += "+env"
		}
	}
}

func validateStrict(cfg *Config) error {
	strict := os.Getenv("BACKEND_REBUILD_STRICT_CONFIG")
	if strict != "1" && strict != "true" && strict != "TRUE" {
		return nil
	}
	if cfg.JWTSecret == "" {
		return fmt.Errorf("strict config: BACKEND_REBUILD_JWT_SECRET (or jwt.secret) is required")
	}
	if cfg.MySQLDSN == "" {
		return fmt.Errorf("strict config: BACKEND_REBUILD_MYSQL_DSN (or mysql.dsn) is required")
	}
	return nil
}
