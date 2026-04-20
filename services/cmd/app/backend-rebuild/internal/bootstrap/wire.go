// Layer: Bootstrap (启动/依赖注入层)
// Responsibility: 组装各层依赖、初始化数据库连接、构建路由、服务编排
// Dependency: 依赖所有层，是系统的组合根(Composition Root)
package bootstrap

import (
	"FeasOJ/app/backend-rebuild/internal/config"
	"FeasOJ/app/backend-rebuild/internal/http/handler"
	"FeasOJ/app/backend-rebuild/internal/http/router"
	"FeasOJ/app/backend-rebuild/internal/queue"
	adminrepo "FeasOJ/app/backend-rebuild/internal/repository/admin"
	authrepo "FeasOJ/app/backend-rebuild/internal/repository/auth"
	classesrepo "FeasOJ/app/backend-rebuild/internal/repository/classes"
	competitionsrepo "FeasOJ/app/backend-rebuild/internal/repository/competitions"
	discussionsrepo "FeasOJ/app/backend-rebuild/internal/repository/discussions"
	problemsrepo "FeasOJ/app/backend-rebuild/internal/repository/problems"
	submitrecordsrepo "FeasOJ/app/backend-rebuild/internal/repository/submitrecords"
	testcasesrepo "FeasOJ/app/backend-rebuild/internal/repository/testcases"
	usersrepo "FeasOJ/app/backend-rebuild/internal/repository/users"
	"FeasOJ/app/backend-rebuild/internal/scheduler"
	adminusecase "FeasOJ/app/backend-rebuild/internal/usecase/admin"
	authusecase "FeasOJ/app/backend-rebuild/internal/usecase/auth"
	classesusecase "FeasOJ/app/backend-rebuild/internal/usecase/classes"
	competitionsusecase "FeasOJ/app/backend-rebuild/internal/usecase/competitions"
	discussionsusecase "FeasOJ/app/backend-rebuild/internal/usecase/discussions"
	problemsusecase "FeasOJ/app/backend-rebuild/internal/usecase/problems"
	"FeasOJ/app/backend-rebuild/internal/usecase/stub"
	submitrecordsusecase "FeasOJ/app/backend-rebuild/internal/usecase/submitrecords"
	testcasesusecase "FeasOJ/app/backend-rebuild/internal/usecase/testcases"
	usersusecase "FeasOJ/app/backend-rebuild/internal/usecase/users"
	"context"
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
			schedulerCtx := context.Background()
			authRepository := authrepo.NewUserRepository(db)
			usersRepository := usersrepo.NewUserRepository(db)
			problemsRepository := problemsrepo.NewProblemRepository(db)
			adminRepository := adminrepo.NewUserRepository(db)
			classesRepository := classesrepo.NewRepository(db)
			competitionsRepository := competitionsrepo.NewRepository(db)
			discussionsRepository := discussionsrepo.NewRepository(db)
			submitRecordsRepository := submitrecordsrepo.NewRepository(db)
			testcasesRepository := testcasesrepo.NewRepository(db)

			submissionQueue := queue.NewMemorySubmissionQueue()
			if cfg.EnableRabbitMQQueue && cfg.RabbitMQURL != "" {
				if rabbitQueue, err := queue.NewRabbitMQSubmissionQueue(queue.RabbitMQConfig{
					URL:        cfg.RabbitMQURL,
					Exchange:   cfg.RabbitMQExchange,
					MainQueue:  cfg.RabbitMQMainQueue,
					RetryQueue: cfg.RabbitMQRetryQueue,
					DLQ:        cfg.RabbitMQDLQ,
					Prefetch:   cfg.RabbitMQWorkerPrefetch,
				}); err != nil {
					log.Printf("[backend-rebuild] rabbitmq init failed, fallback to memory queue: %v", err)
				} else {
					submissionQueue = rabbitQueue
					log.Println("[backend-rebuild] rabbitmq submission queue enabled")
				}
			}

			services.Auth = authusecase.NewService(authRepository, cfg.JWTSecret, cfg.JWTIssuer, jwtTTL)
			services.Users = usersusecase.NewService(usersRepository)
			services.Problems = problemsusecase.NewService(problemsRepository)
			services.Admin = adminusecase.NewService(adminRepository)
			services.Classes = classesusecase.NewService(classesRepository)
			services.Competitions = competitionsusecase.NewService(competitionsRepository)
			services.Discussions = discussionsusecase.NewService(discussionsRepository)
			services.SubmitRecords = submitrecordsusecase.NewService(submitRecordsRepository, submissionQueue)
			services.Testcases = testcasesusecase.NewService(testcasesRepository)
			if cfg.EnableEmbeddedJudgeWorker && cfg.EnableJudgeWriteback {
				queue.NewJudgeWorker(submissionQueue, services.SubmitRecords, 300*time.Millisecond).Start(schedulerCtx)
			}
			scheduler.StartContestStatusReconciler(schedulerCtx, db, time.Duration(cfg.ContestStatusScanSeconds)*time.Second)
			log.Println("[backend-rebuild] all phase-1 services enabled with mysql backend")
		}
	}

	h := handler.New(handler.Handlers{
		Auth:          services.Auth,
		Users:         services.Users,
		Classes:       services.Classes,
		Problems:      services.Problems,
		Testcases:     services.Testcases,
		Competitions:  services.Competitions,
		Discussions:   services.Discussions,
		SubmitRecords: services.SubmitRecords,
		Admin:         services.Admin,
		JudgeWritebackToken: cfg.JudgeWritebackToken,
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
