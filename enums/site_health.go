package enums

import (
	"fmt"
	"strings"
)

type SiteHealth string

const (
	Pending   SiteHealth = "pending"
	Healthy   SiteHealth = "healthy"
	UnHealthy SiteHealth = "unhealthy"
)

func ParseSiteHealth(s string) (SiteHealth, error) {
	switch strings.ToLower(s) {
	case "pending":
		return Pending, nil
	case "healthy":
		return Healthy, nil
	case "unhealthy":
		return UnHealthy, nil
	default:
		return "", fmt.Errorf("invalid site health option: %s", s)
	}
}
