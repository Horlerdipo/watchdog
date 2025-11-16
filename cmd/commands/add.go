package commands

import (
	"context"
)

type AddCommand struct {
}

func (mc AddCommand) Name() string {
	return "add"
}

func (mc AddCommand) Action(ctx context.Context, cmd CommandContext) error {
	//add it to db
	//trigger reentry into the redis list
	return nil
}

func (mc AddCommand) Aliases() []string {
	return []string{"a"}
}

func (mc AddCommand) Usage() string {
	return "Add a URL to the watchdog monitoring process."
}

func NewAddCommand() *AddCommand {
	return &AddCommand{}
}
