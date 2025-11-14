package migrations

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"

	"mPR/internal/config"
)

func Run(cfg config.Database, log *zap.Logger) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.Mode,
	)

	migrations, err := migrate.New("file://db/scripts", dsn)
	if err != nil {
		log.Panic("Error run migrations", zap.Error(err))
		return
	}
	defer func() { _, _ = migrations.Close() }()

	if err := migrations.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Info("Error application migrations", zap.Error(err))
		return
	}
}
