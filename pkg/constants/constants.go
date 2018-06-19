package constants

import (
	"time"
)

// Constants for environment variable names.
const (
	EnvGoPXVCSAPIAuthUser     = "GOPX_VCS_API_AUTH_USER"
	EnvGoPXVSCAPIAuthPassword = "GOPX_VCS_API_AUTH_PASSWORD"
)

// Constants for timeouts during server request processing.
const (
	ServerReadTimeout  = time.Second * 15
	ServerWriteTimeout = time.Second * 15
	ServerIdleTimeout  = time.Second * 15
)
