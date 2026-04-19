// Layer: Bootstrap (启动/依赖注入层)
// Responsibility: 组装各层依赖、初始化数据库连接、构建路由、服务编排
// Dependency: 依赖所有层，是系统的组合根(Composition Root)
package bootstrap

import (
	"FeasOJ/app/backend-rebuild/internal/config"
	"FeasOJ/app/backend-rebuild/internal/http/handler"
	"FeasOJ/app/backend-rebuild/internal/http/router"
	adminrepo "FeasOJ/app/backend-rebuild/internal/repository/admin"
	authrepo "FeasOJ/app/backend-rebuild/internal/repository/auth"
	problemsrepo "FeasOJ/app/backend-rebuild/internal/repository/problems"
	usersrepo "FeasOJ/app/backend-rebuild/internal/repository/users"
	adminusecase "FeasOJ/app/backend-rebuild/internal/usecase/admin"
	authusecase "FeasOJ/app/backend-rebuild/internal/usecase/auth"
	problemsusecase "FeasOJ/app/backend-rebuild/internal/usecase/problems"
	"FeasOJ/app/backend-rebuild/internal/usecase/stub"
	usersusecase "FeasOJ/app/backend-rebuild/internal/usecase/users"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func BuildRouter(cfg config.Config) *gin.Engine {
	g := gin.New()
	g.Use(gin.Recovery())

	services := stub.NewServices()
	jwtTTL := parseJWTTTL(cfg.JWTExpireH)
	if cfg.JWTSecret == "" {
		log.Println("[backend-rebuild] BACKEND_REBUILD_JWT_SECRET is empty, using unsafe dev default")
		cfg.JWTSecret = "dev-secret-change-me"
	}

	if cfg.MySQLDSN == "" {
		log.Println("[backend-rebuild] BACKEND_REBUILD_MYSQL_DSN is empty, auth service stays in stub mode")
	} else {
		db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
		if err != nil {
			log.Printf("[backend-rebuild] mysql connect failed, auth service stays in stub mode: %v", err)
		} else {
			authRepository := authrepo.NewUserRepository(db)
			usersRepository := usersrepo.NewUserRepository(db)
			problemsRepository := problemsrepo.NewProblemRepository(db)
			adminRepository := adminrepo.NewUserRepository(db)
			services.Auth = authusecase.NewService(authRepository, cfg.JWTSecret, cfg.JWTIssuer, jwtTTL)
			services.Users = usersusecase.NewService(usersRepository)
			services.Problems = problemsusecase.NewService(problemsRepository)
			services.Admin = adminusecase.NewService(adminRepository)
			log.Println("[backend-rebuild] auth/users/problems/admin services enabled with mysql backend")
		}
	}

	h := handler.New(handler.Handlers{
		Auth:          services.Auth,
		Users:         services.Users,
		Classes:       services.Classes,
		Problems:      services.Problems,
		Competitions:  services.Competitions,
		Discussions:   services.Discussions,
		SubmitRecords: services.SubmitRecords,
		Admin:         services.Admin,
	})

	router.Register(g, h, cfg)
	return g
}

func parseJWTTTL(hoursText string) time.Duration {
	hours, err := strconv.Atoi(hoursText)
	if err != nil || hours <= 0 {
		hours = 72
	}
	return time.Duration(hours) * time.Hour
}
