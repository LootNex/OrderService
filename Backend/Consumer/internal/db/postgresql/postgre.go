package postgresql

import (
	"database/sql"
	"fmt"

	"github.com/LootNex/OrderService/Consumer/configs"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func InitPostgres(config *config.Config, log *zap.Logger) (*sql.DB, error) {

	strConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Postgres.Host, config.Postgres.Port, config.Postgres.User, config.Postgres.Password, config.Postgres.DBname)

	db, err := sql.Open("postgres", strConn)
	if err != nil {
		return nil, fmt.Errorf("cannot open posgres err:%w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping err:%w", err)
	}
	log.Info("Postgres is running")

	if err = RunMigrations(db, log); err != nil {
		return nil, err
	}

	return db, nil
}
