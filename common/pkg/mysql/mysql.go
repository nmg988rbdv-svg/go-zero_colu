package mysql

import (
	"fmt"
	"time"

	logger "github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/zlog"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

// Conf describes a MySQL connection shared by the account and quote services.
type Conf struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	Charset         string        `json:"charset,optional"`
	MaxIdleConns    int           `json:"maxIdleConns,optional"`
	MaxOpenConns    int           `json:"maxOpenConns,optional"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime,optional"`
	Logger          logger.Config `json:"logger,optional"`
}

func (c Conf) dsn() string {
	charset := c.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, charset)
}

// MustNewClient opens and verifies a MySQL connection. Startup fails fast when
// the database is unavailable because all durable exchange data depends on it.
func (c Conf) MustNewClient() *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.dsn()), c.gormConfig())
	if err != nil {
		logx.Severef("connect mysql failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logx.Severef("get mysql sql.DB failed: %v", err)
	}
	if c.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	}
	if c.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	}
	if c.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(c.ConnMaxLifetime)
	}
	if err := sqlDB.Ping(); err != nil {
		logx.Severef("ping mysql failed: %v", err)
	}
	return db
}

func (c Conf) gormConfig() *gorm.Config {
	config := &gorm.Config{SkipDefaultTransaction: true}

	var l *zap.Logger
	if c.Logger.Mode == "" {
		l = logger.GetZapLogger()
	} else {
		l = c.Logger.Build()
	}
	gl := zapgorm2.New(l.WithOptions(zap.AddCallerSkip(1)))
	gl.IgnoreRecordNotFoundError = true
	gl.LogLevel = gormlogger.Warn
	config.Logger = gl
	return config
}
