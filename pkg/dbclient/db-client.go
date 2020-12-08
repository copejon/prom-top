/*
Copyright 2020 Red Hat, Inc. jcope@redhat.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package dbclient

import (
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

// The column names expected to exist in the database table
const (
	// Build names the server version
	Build      string = "build"
	// Metric storage the prometheus model.Metric, a series of key=value labels
	Metric     string = "metric"
	// Value is the returned scalar value of the metric
	Value      string = "value"
	// QueryTime records the time and data the query we executed
	QueryTime string = "query_time"

	// Table is a hardcoded table name.  Will be replaced with dynamically set names.
	Table string = "metrics"
)

const (
	host     = "PGHOST"
	port     = "PGPORT"
	database = "PGDATABASE"
	user     = "PGUSER"
	password = "PGPASSWORD"
)



type PostgresConfig struct {
	host     string
	port     int
	database string
	user     string
	password string
}

func (p PostgresConfig) String () string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", p.user, p.password, p.host, p.port, p.database)
}

func initConfig() PostgresConfig {
	exPath, _ := os.Executable()
	viper.SetConfigFile(filepath.Join(filepath.Dir(exPath), ".env"))
	viper.SetConfigType("dotenv")
	err := viper.BindEnv(
		host,
		port,
		database,
		user,
		password)
	if err != nil {
		log.Fatalf("failed to bind env vars: %v", err)
	}
	viper.AutomaticEnv()
	err = viper.ReadInConfig()

	if err != nil {
		log.Printf("no postgress config found: %v", err)
	}

	return PostgresConfig{
		host : viper.GetString(host),
		port : viper.GetInt(port),
		database : viper.GetString(database),
		user : viper.GetString(user),
		password : viper.GetString(password),
	}
}

func NewPostgresClient() (*sqlx.DB, error) {
	cfg := initConfig()
	return sqlx.Connect("pgx", cfg.String())
}