package postgres

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mPR/internal/config"
)

func New(cfg config.Database, log *zap.Logger) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.Username, cfg.Password, cfg.Name, cfg.Port, cfg.Mode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("Ошибка подключения к бд", zap.Error(err))
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Panic("Ошибка получения sqlDB", zap.Error(err))
		return nil
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	return db
}
