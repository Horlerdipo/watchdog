package enums

import "fmt"

type MonitoringFrequency string

const (
	TenSeconds      MonitoringFrequency = "ten_seconds"
	ThirtySeconds   MonitoringFrequency = "thirty_seconds"
	OneMinute       MonitoringFrequency = "one_minute"
	FiveMinutes     MonitoringFrequency = "five_minutes"
	ThirtyMinute    MonitoringFrequency = "thirty_minutes"
	OneHour         MonitoringFrequency = "one_hour"
	TwelveHours     MonitoringFrequency = "twelve_hours"
	TwentyFourHours MonitoringFrequency = "twenty_four_hours"
)

func (m MonitoringFrequency) ToSeconds() int {
	switch m {
	case TenSeconds:
		return 10
	case ThirtySeconds:
		return 30
	case OneMinute:
		return 60
	case FiveMinutes:
		return 300
	case ThirtyMinute:
		return 1800
	case OneHour:
		return 3600
	case TwelveHours:
		return 43200
	case TwentyFourHours:
		return 86400
	default:
		return -1 // or panic/error
	}
}

func ParseMonitoringFrequency(s string) (MonitoringFrequency, error) {
	switch s {
	case "ten_seconds":
		return TenSeconds, nil
	case "thirty_seconds":
		return ThirtySeconds, nil
	case "one_minute":
		return OneMinute, nil
	case "five_minutes":
		return FiveMinutes, nil
	case "thirty_minutes":
		return ThirtyMinute, nil
	case "one_hour":
		return OneHour, nil
	case "twelve_hours":
		return TwelveHours, nil
	case "twenty_four_hours":
		return TwentyFourHours, nil
	default:
		return "", fmt.Errorf("invalid monitoring frequency: %s", s)
	}
}
