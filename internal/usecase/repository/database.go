package repository

import (
	"cernunnos/internal/pkg/config"
	sqlutils "cernunnos/internal/pkg/sqltools"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func ProvideDatabaseConnection(c *config.Config) (*sql.DB, func(), error) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s/cernunnos?sslmode=disable",
		c.DatabaseUser, c.DatabasePassword, c.DatabaseHost)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, func() {}, fmt.Errorf("error connecting to database: %w", err)
	}

	return db, func() {
		db.Close()
	}, nil
}

type Repository interface {
	DB() sqlutils.DBTX
}
