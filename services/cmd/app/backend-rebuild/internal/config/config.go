// Layer: Config (配置层)
// Responsibility: 配置加载(YAML+环境变量)、默认值管理、配置聚合
// Dependency: 不依赖其他层，被 Bootstrap 层调用
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

const defaultAddr = "127.0.0.1:8082"
const defaultJWTIssuer = "feasoj-backend-rebuild"
const defaultJWTExpireHours = "2"
const defaultContestStatusScanSeconds = 30
const defaultConfigPath = "app/backend-rebuild/config.yaml"
const defaultAvatarUploadDir = "app/backend-rebuild/var/avatars"
const defaultAvatarUploadMaxBytes int64 = 2 * 1024 * 1024

type Config struct {
	ConfigPath                string
	ConfigSource              string
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
	EnableAvatarUpload        bool
	RabbitMQURL               string
	RabbitMQExchange          string
	RabbitMQMainQueue         string
	RabbitMQRetryQueue        string
	RabbitMQDLQ               string
	RabbitMQWorkerPrefetch    int
	AvatarUploadDir           string
	AvatarUploadMaxBytes      int64
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
		EnableAvatarUpload        *bool `yaml:"enable_avatar_upload"`
	} `yaml:"feature_flags"`
	AvatarStorage struct {
		Dir      string `yaml:"dir"`
		MaxBytes int64  `yaml:"max_bytes"`
	} `yaml:"avatar_storage"`
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
		EnableTestcaseAPIs:        true,
		EnableRabbitMQQueue:       true,
		EnableEmbeddedJudgeWorker: true,
		EnableClassWorkflowV2:     true,
		EnablePasswordReset:       false,
		EnableAvatarUpload:        true,
		RabbitMQExchange:          "judge.submission.exchange",
		RabbitMQMainQueue:         "judge.submission.main",
		RabbitMQRetryQueue:        "judge.submission.retry",
		RabbitMQDLQ:               "judge.submission.dlq",
		RabbitMQWorkerPrefetch:    1,
		AvatarUploadDir:           defaultAvatarUploadDir,
		AvatarUploadMaxBytes:      defaultAvatarUploadMaxBytes,
	}

	if err := loadFromFile(&cfg); err != nil {
		return Config{}, err
	}
	loadFromEnv(&cfg)
	if err := validateStrict(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadFromFile(cfg *Config) error {
	path := os.Getenv("BACKEND_REBUILD_CONFIG")
	if path == "" {
		path = defaultConfigPath
	}
	cfg.ConfigPath = path

	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" || ext == ".yaml" || ext == ".yml" {
		return loadFromYAML(cfg, path)
	}
	if ext == ".toml" {
		if !compatModeEnabled() {
			return fmt.Errorf("config format %q is not supported when BACKEND_REBUILD_CONFIG_COMPAT_MODE=false; migrate to YAML", ext)
		}
		log.Printf("[backend-rebuild] deprecated config format TOML accepted in compatibility mode, please migrate %s to YAML", path)
		return loadFromTOML(cfg, path)
	}

	return fmt.Errorf("unsupported config extension %q, only .yaml/.yml are supported (or .toml in compatibility mode)", ext)
}

func compatModeEnabled() bool {
	v := strings.TrimSpace(os.Getenv("BACKEND_REBUILD_CONFIG_COMPAT_MODE"))
	if v == "" {
		return true
	}
	return v == "1" || strings.EqualFold(v, "true")
}

func loadFromYAML(cfg *Config, path string) error {
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
	if fc.FeatureFlags.EnableAvatarUpload != nil {
		cfg.EnableAvatarUpload = *fc.FeatureFlags.EnableAvatarUpload
	}
	if fc.AvatarStorage.Dir != "" {
		cfg.AvatarUploadDir = fc.AvatarStorage.Dir
	}
	if fc.AvatarStorage.MaxBytes > 0 {
		cfg.AvatarUploadMaxBytes = fc.AvatarStorage.MaxBytes
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

func loadFromTOML(cfg *Config, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.ConfigSource = "defaults"
			return nil
		}
		return fmt.Errorf("read config toml failed: %w", err)
	}

	var fc struct {
		Server struct {
			Addr string `toml:"addr"`
		} `toml:"server"`
		MySQL struct {
			DSN string `toml:"dsn"`
		} `toml:"mysql"`
		JWT struct {
			Secret      string `toml:"secret"`
			Issuer      string `toml:"issuer"`
			ExpireHours int    `toml:"expire_hours"`
		} `toml:"jwt"`
		Judge struct {
			WritebackToken string `toml:"writeback_token"`
		} `toml:"judge"`
		Scheduler struct {
			ContestStatusScanSeconds int `toml:"contest_status_scan_seconds"`
		} `toml:"scheduler"`
		FeatureFlags struct {
			EnableJudgeWriteback      *bool `toml:"enable_judge_writeback"`
			EnableScoreboard          *bool `toml:"enable_scoreboard"`
			EnableTestcaseAPIs        *bool `toml:"enable_testcase_apis"`
			EnableRabbitMQQueue       *bool `toml:"enable_rabbitmq_queue"`
			EnableEmbeddedJudgeWorker *bool `toml:"enable_embedded_judge_worker"`
			EnableClassWorkflowV2     *bool `toml:"enable_class_workflow_v2"`
			EnablePasswordReset       *bool `toml:"enable_password_reset"`
			EnableAvatarUpload        *bool `toml:"enable_avatar_upload"`
		} `toml:"feature_flags"`
		AvatarStorage struct {
			Dir      string `toml:"dir"`
			MaxBytes int64  `toml:"max_bytes"`
		} `toml:"avatar_storage"`
		RabbitMQ struct {
			URL        string `toml:"url"`
			Exchange   string `toml:"exchange"`
			MainQueue  string `toml:"main_queue"`
			RetryQueue string `toml:"retry_queue"`
			DLQ        string `toml:"dlq"`
			Prefetch   int    `toml:"worker_prefetch"`
		} `toml:"rabbitmq"`
	}

	if _, err := toml.Decode(string(content), &fc); err != nil {
		return fmt.Errorf("parse config toml failed: %w", err)
	}
	cfg.ConfigSource = "toml-compat:" + path

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
	if fc.FeatureFlags.EnableAvatarUpload != nil {
		cfg.EnableAvatarUpload = *fc.FeatureFlags.EnableAvatarUpload
	}
	if fc.AvatarStorage.Dir != "" {
		cfg.AvatarUploadDir = fc.AvatarStorage.Dir
	}
	if fc.AvatarStorage.MaxBytes > 0 {
		cfg.AvatarUploadMaxBytes = fc.AvatarStorage.MaxBytes
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
	if v := os.Getenv("BACKEND_REBUILD_ENABLE_AVATAR_UPLOAD"); v != "" {
		cfg.EnableAvatarUpload = v == "1" || v == "true" || v == "TRUE"
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_AVATAR_UPLOAD_DIR"); v != "" {
		cfg.AvatarUploadDir = v
		envOverride = true
	}
	if v := os.Getenv("BACKEND_REBUILD_AVATAR_UPLOAD_MAX_BYTES"); v != "" {
		envOverride = true
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			cfg.AvatarUploadMaxBytes = n
		}
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
	if cfg.AvatarUploadDir == "" {
		cfg.AvatarUploadDir = defaultAvatarUploadDir
	}
	if cfg.AvatarUploadMaxBytes <= 0 {
		cfg.AvatarUploadMaxBytes = defaultAvatarUploadMaxBytes
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
