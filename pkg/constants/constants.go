package constants

import (
	"time"
)

// Constants related to Database
var (
	DbConnMaxLifeTime = time.Hour * 10
	DbMaxIdleConn     = 10
	DbMaxOpenConn     = 10
)
