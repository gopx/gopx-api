package database

import (
	"database/sql"
	"net"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"gopx.io/gopx-api/pkg/config"
	"gopx.io/gopx-api/pkg/constants"
	"gopx.io/gopx-common/log"
)

var dsnConfig = mysql.Config{
	Net:                     "tcp",
	Addr:                    net.JoinHostPort(config.Db.Host, strconv.Itoa(config.Db.Port)),
	User:                    config.Db.Username,
	Passwd:                  config.Db.Password,
	DBName:                  config.Db.DatabaseName,
	Loc:                     time.Local,
	ParseTime:               true,
	Params:                  nil,
	Collation:               "utf8_general_ci",
	MaxAllowedPacket:        4194304,
	ServerPubKey:            "",
	TLSConfig:               "",
	Timeout:                 0,
	ReadTimeout:             0,
	WriteTimeout:            0,
	AllowAllFiles:           false,
	AllowCleartextPasswords: false,
	AllowNativePasswords:    true,
	AllowOldPasswords:       false,
	ClientFoundRows:         false,
	ColumnsWithAlias:        false,
	InterpolateParams:       false,
	MultiStatements:         true,
	RejectReadOnly:          false,
}

var dbConn *sql.DB

func init() {
	db, err := sql.Open("mysql", dsnConfig.FormatDSN())
	if err != nil {
		log.Fatal("Error: %s", err)
	}

	db.SetConnMaxLifetime(constants.DbConnMaxLifeTime)
	db.SetMaxIdleConns(constants.DbMaxIdleConn)
	db.SetMaxOpenConns(constants.DbMaxOpenConn)

	dbConn = db
}

// Conn returns the database connection pool instance.
func Conn() *sql.DB {
	return dbConn
}
