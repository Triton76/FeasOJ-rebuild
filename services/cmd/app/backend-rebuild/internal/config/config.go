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
const defaultJWTExpireHours = "72"
const defaultContestStatusScanSeconds = 30
const defaultConfigPath = "app/backend-rebuild/config.yaml"

type Config struct {
	Addr                     string
	MySQLDSN                 string
	JWTSecret                string
	JWTIssuer                string
	JWTExpireH               string
	ContestStatusScanSeconds int
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
	Scheduler struct {
		ContestStatusScanSeconds int `yaml:"contest_status_scan_seconds"`
	} `yaml:"scheduler"`
}

func Load() (Config, error) {
	cfg := Config{
		Addr:                     defaultAddr,
		JWTIssuer:                defaultJWTIssuer,
		JWTExpireH:               defaultJWTExpireHours,
		ContestStatusScanSeconds: defaultContestStatusScanSeconds,
	}

	if err := loadFromYAML(&cfg); err != nil {
		return Config{}, err
	}
	loadFromEnv(&cfg)

	return cfg, nil
}

func loadFromYAML(cfg *Config) error {
	path := os.Getenv("BACKEND_REBUILD_CONFIG")
	if path == "" {
		path = defaultConfigPath
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config yaml failed: %w", err)
	}

	var fc fileConfig
	if err := yaml.Unmarshal(content, &fc); err != nil {
		return fmt.Errorf("parse config yaml failed: %w", err)
	}

	if fc.Server.Addr != "" {
		cfg.Addr = fc.Server.Addr
	}
	if fc.MySQL.DSN != "" {
		cfg.MySQLDSN = fc.MySQL.DSN
	}
	if fc.JWT.Secret != "" {
		cfg.JWTSecret = fc.JWT.Secret
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

	return nil
}

func loadFromEnv(cfg *Config) {
	if v := os.Getenv("BACKEND_REBUILD_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("BACKEND_REBUILD_MYSQL_DSN"); v != "" {
		cfg.MySQLDSN = v
	}
	if v := os.Getenv("BACKEND_REBUILD_JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("BACKEND_REBUILD_JWT_ISSUER"); v != "" {
		cfg.JWTIssuer = v
	}
	if v := os.Getenv("BACKEND_REBUILD_JWT_EXPIRE_HOURS"); v != "" {
		cfg.JWTExpireH = v
	}
	if v := os.Getenv("BACKEND_REBUILD_CONTEST_STATUS_SCAN_SECONDS"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			cfg.ContestStatusScanSeconds = sec
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
}
