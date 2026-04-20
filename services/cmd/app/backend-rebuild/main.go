// Layer: Main (应用入口层)
// Responsibility: 程序入口、配置加载、启动服务
// Dependency: 依赖 Config 和 Bootstrap 层，是整个应用的启动点
package main

import (
	"FeasOJ/app/backend-rebuild/internal/bootstrap"
	"FeasOJ/app/backend-rebuild/internal/config"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[backend-rebuild] load config failed: %v", err)
	}
	log.Printf("[backend-rebuild] config source=%s path=%s", cfg.ConfigSource, cfg.ConfigPath)

	r := bootstrap.BuildRouter(cfg)
	log.Printf("[backend-rebuild] listening on %s", cfg.Addr)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("[backend-rebuild] server exited: %v", err)
	}
}
