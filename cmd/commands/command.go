package commands

import (
	"context"
	"github.com/urfave/cli/v3"
)

var AvailableCommands = map[string]Command{}

type Command interface {
	Name() string
	Action(ctx context.Context, cmd *cli.Command) error
	Aliases() []string
	Usage() string
}

type CommandContainer struct {
	Commands []Command
}

func (cc *CommandContainer) Register(command Command) {
	cc.Commands = append(cc.Commands, command)
	return
}

func (cc *CommandContainer) RegisterAll() {
	guardCommand := NewGuardCommand()
	cc.Register(guardCommand)
}

func (cc *CommandContainer) Initiate() []*cli.Command {
	cc.RegisterAll()
	var commands []*cli.Command
	for _, command := range cc.Commands {
		commands = append(commands, &cli.Command{
			Name:    command.Name(),
			Usage:   command.Usage(),
			Aliases: command.Aliases(),
			Action:  command.Action,
		})
	}
	return commands
}
