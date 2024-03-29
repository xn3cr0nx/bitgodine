package postgres

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
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
	*gorm.DB
}

// Conf defines needed field for connecting to Postgres instance
func Conf() *Config {
	return &Config{
		Host:    viper.GetString("postgres.host"),
		Port:    viper.GetInt("postgres.port"),
		Pass:    viper.GetString("postgres.pass"),
		User:    viper.GetString("postgres.user"),
		SSLMode: viper.GetString("postgres.sslmode"),
		Name:    viper.GetString("postgres.name"),
		Debug:   viper.GetBool("postgres.debug"),
	}
}

// NewPg returns a new instance of postgres db. The configuration need to be correct
// in order to enable postgres to be connected with Connect on receiver method
func NewPg(conf *Config) (*Pg, error) {
	if conf.Host == "" {
		return &Pg{}, fmt.Errorf("%w: missing or invalid host", errorx.ErrConfig)
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

	conStr := conf.String()
	level := gormLogger.Error
	if conf.Debug {
		level = gormLogger.Info
	}

	db, err := gorm.Open(postgres.Open(conStr), &gorm.Config{
		Logger: gormLogger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}
	return &Pg{db}, nil
}

// String returns the connection string for connecting to postgres
func (c *Config) String() string {
	strBase := "host=%s port=%s user=%s dbname=%s password=%s sslmode=%s"
	return fmt.Sprintf(strBase, c.Host, strconv.Itoa(c.Port), c.User, c.Name, c.Pass, c.SSLMode)
}

// MigrationString returns the connection string in url format for connecting to postgres
func (c *Config) MigrationString() string {
	strBase := "postgres://%s:%s@%s:%s/%s?sslmode=%s"
	return fmt.Sprintf(strBase, c.User, c.Pass, c.Host, strconv.Itoa(c.Port), c.Name, c.SSLMode)
}

// // repository interface compatibility
// func wrap(db *gorm.DB) *Pg {
// 	return &Pg{db}
// }

// // Create wrapper form gorm DB wrapper
// func (pg *Pg) Create(value interface{}) *Pg {
// 	return wrap(pg.DB.Create(value))
// }

// // Model wrapper form gorm DB wrapper
// func (pg *Pg) Model(value interface{}) *Pg {
// 	return wrap(pg.DB.Model(value))
// }

// // Find wrapper form gorm DB wrapper
// func (pg *Pg) Find(dest interface{}, conds ...interface{}) *Pg {
// 	return wrap(pg.DB.Find(dest, conds))
// }

// // Where wrapper form gorm DB wrapper
// func (pg *Pg) Where(query interface{}, conds ...interface{}) *Pg {
// 	return wrap(pg.DB.Where(query, conds))
// }

// // Error returns query error
// func (pg *Pg) Error() error {
// 	return pg.DB.Error
// }
