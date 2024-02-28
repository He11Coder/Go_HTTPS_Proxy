package repository

import (
	"database/sql"

	_ "github.com/jackc/pgx/stdlib"
)

type postgresConfig struct {
	user     string
	password string
	dbname   string
	host     string
	port     string
	sslmode  string
}

var requestPostgresConfig = postgresConfig{
	user:     "admin",
	password: "password",
	dbname:   "request",
	host:     "host.docker.internal",
	port:     "8055",
	sslmode:  "disable",
}

func (conf postgresConfig) GetConnectionString() string {
	return "user=" + conf.user + " password=" + conf.password + " dbname=" + conf.dbname +
		" host=" + conf.host + " port=" + conf.port +
		" sslmode=" + conf.sslmode
}

func GetPostgres() (*sql.DB, error) {
	dsn := requestPostgresConfig.GetConnectionString()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, err
}
