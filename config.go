package downthedrive

import (
	"time"
)

type DownTheDriveConfig struct {
	Authentication DownTheDriveAuthenticationConfig
	Db             DownTheDriveDbConfig
	IndexStrategy  IndexStrategy
	ItemsMaxAge    time.Duration
}

type DownTheDriveAuthenticationConfig struct {
	ClientId     string
	ClientSecret string
}

type DownTheDriveDbConfig struct {
	DataSource string
	Driver     string
}
