package listeners

import (
	"fmt"
	"github.com/horlerdipo/watchdog/core"
	"github.com/horlerdipo/watchdog/events"
)

type PingSuccessfulListener struct {
}

func (sl *PingSuccessfulListener) Handle(event core.Event) {
	e := event.(*events.PingSuccessful)
	fmt.Printf("%v is healthy, pushing to timescale DB \n", e.Url)
}

func NewPingSuccessfulListener() *PingSuccessfulListener {
	return &PingSuccessfulListener{}
}
