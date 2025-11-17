package database

import (
	"github.com/horlerdipo/watchdog/enums"
	"time"
)

type Url struct {
	Id                  int
	Url                 string
	HttpMethod          enums.HttpMethod
	Status              enums.SiteHealth
	MonitoringFrequency enums.MonitoringFrequency
	ContactEmail        string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
