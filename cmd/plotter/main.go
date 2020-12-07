package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type Row struct {
	Build     string
	Metric    string
	Value     float64
	QueryTime time.Time `db:"query_time"`
}

func (r Row) String() string {
	return fmt.Sprintf("Ver: %s | Metric: %s | Val: %f | Time: %v", r.Build, r.Metric, r.Value, r.QueryTime)
}

const (
	PGHOST     = `PGHOST`
	PGPORT     = `PGPORT`
	PGDATABASE = `PGDATABASE`
	PGUSER     = `PGUSER`
	PGPASSWORD = `PGPASSWORD`
)

type config struct {
	host     string
	port     int
	database string
	user     string
	password string
}

func (c config) ToDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.user, c.password, c.host, c.port, c.database)
}

func main() {
	base, _ := os.Executable()
	viper.SetConfigFile(filepath.Join(filepath.Dir(base), ".env"))
	viper.SetConfigType("dotenv")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("read config file error: %v", err)
	}

	cfg := config{
		host:     viper.GetString(PGHOST),
		port:     viper.GetInt(PGPORT),
		database: viper.GetString(PGDATABASE),
		user:     viper.GetString(PGUSER),
		password: viper.GetString(PGPASSWORD),
	}

	conn, err := sqlx.Connect("pgx", cfg.ToDSN())
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer conn.Close()

	rows := make([]Row, 0)
	err = conn.Select(&rows, "SELECT * FROM metrics")
	if err != nil {
		log.Fatalf("select failed: %v", err)
	}
	log.Printf("got %d rows\n", len(rows))
	for _, r := range rows {
		log.Println("Row --- " + r.String())
	}
}
