package postgres

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Config instance of Postgres configuration details
type Config struct {
	Host    string
	Port    int
	SSLMode string
	Name    string
	User    string
	Pass    string
	Debug   bool
}

// Pg instance of Postgres db with configuration details
type Pg struct {
	DB   *gorm.DB
	conf *Config
}

// Conf defines needed field for connecting to Postgres instance
func Conf() *Config {
	return &Config{
		Host:    viper.GetString("db.postgres.host"),
		Port:    viper.GetInt("db.postgres.port"),
		Pass:    viper.GetString("db.postgres.pass"),
		User:    viper.GetString("db.postgres.user"),
		SSLMode: viper.GetString("db.postgres.sslmode"),
		Name:    viper.GetString("db.postgres.name"),
		Debug:   viper.GetBool("db.postgres.debug"),
	}
}

// NewPg returns a new instance of postgres db. The configuration need to be correct
// in order to enable postgres to be connected with Connect on receiver method
func NewPg(conf *Config) (*Pg, error) {
	if conf.Host == "" {
		return &Pg{}, errors.New("Missing or invalid host")
	}
	if conf.Port == 0 {
		logger.Info("Postgres", "Missing postgres port, using default 5432", logger.Params{})
		conf.Port = 5432
	}
	if conf.Name == "" {
		logger.Info("Postgres", "Missing postgres db name, using default postgres", logger.Params{})
	}
	if conf.SSLMode == "" {
		logger.Info("Postgres", "Missing postgres ssl mode, disabled by default", logger.Params{})
		conf.SSLMode = "disable"
	}
	return &Pg{conf: conf}, nil
}

// String returnes the connection string for connecting to postgres
func (c *Config) String() string {
	strBase := "host=%s port=%s user=%s dbname=%s password=%s sslmode=%s"
	return fmt.Sprintf(strBase, c.Host, strconv.Itoa(c.Port), c.User, c.Name, c.Pass, c.SSLMode)
}

// MigrationString returnes the connection string in url format for connecting to postgres
func (c *Config) MigrationString() string {
	strBase := "postgres://%s:%s@%s:%s/%s?sslmode=%s"
	return fmt.Sprintf(strBase, c.User, c.Pass, c.Host, strconv.Itoa(c.Port), c.Name, c.SSLMode)
}

// Connect open postgres connection using configuration details provided in conf field
func (pg *Pg) Connect() error {
	conStr := pg.conf.String()
	db, err := gorm.Open("postgres", conStr)
	if err != nil {
		return err
	}
	pg.DB = db

	pg.DB.LogMode(pg.conf.Debug)

	return nil
}

// Close interrupts the connection with the postgres instance
func (pg *Pg) Close() {
	pg.DB.Close()
}
