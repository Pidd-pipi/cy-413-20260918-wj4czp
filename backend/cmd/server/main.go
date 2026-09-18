package main

import (
	"fmt"
	"github.com/blueship581/mindgarden/backend/internal/config"
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/handler"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
	"github.com/blueship581/mindgarden/backend/internal/router"
	"github.com/blueship581/mindgarden/backend/internal/service"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log/slog"
	"os"
)

func main() {
	cfg, e := config.Load()
	if e != nil {
		panic(fmt.Errorf("load config: %w", e))
	}
	logger := util.NewLogger()
	db, e := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if e != nil {
		logger.Error("database connect failed", "error", e)
		os.Exit(1)
	}
	logger.Info(constants.LogDBConnected)
	if e = db.AutoMigrate(&model.User{}, &model.Mood{}, &model.Assessment{}, &model.Journal{}, &model.UserAssessment{}); e != nil {
		logger.Error("database migrate failed", "error", e)
		os.Exit(1)
	}
	logger.Info(constants.LogDBMigrated)
	ur := repository.NewUserRepository(db)
	mr := repository.NewMoodRepository(db)
	ar := repository.NewAssessmentRepository(db)
	jr := repository.NewJournalRepository(db)
	us := service.NewUserService(ur, logger)
	ms := service.NewMoodService(mr, logger)
	as := service.NewAssessmentService(ar, logger)
	js := service.NewJournalService(jr, logger)
	if e = as.Seed(); e != nil {
		logger.Error("assessment seed failed", "error", e)
		os.Exit(1)
	}
	h := router.Handlers{User: handler.NewUserHandler(us, as, logger, cfg.JWTSecret, cfg.JWTIssuer), Mood: handler.NewMoodHandler(ms, logger), Assessment: handler.NewAssessmentHandler(as, logger), Journal: handler.NewJournalHandler(js, logger)}
	if e = router.New(cfg, h, logger).Run(":" + cfg.Port); e != nil {
		logger.Error("server stopped", "error", e)
		os.Exit(1)
	}
}

var _ *slog.Logger
