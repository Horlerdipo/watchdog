package commands

import (
	"context"
	"github.com/urfave/cli/v3"
)

type Command interface {
	Name() string
	Action(ctx context.Context, cmd CommandContext) error
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
	cc.Register(NewGuardCommand())
	cc.Register(NewAddCommand())
}

func (cc *CommandContainer) Initiate() []*cli.Command {
	cc.RegisterAll()
	var commands []*cli.Command
	for _, command := range cc.Commands {
		commands = append(commands, &cli.Command{
			Name:    command.Name(),
			Usage:   command.Usage(),
			Aliases: command.Aliases(),
			Action: func(ctx context.Context, cmd *cli.Command) error {
				wrapped := &UrfaveContext{cmd: cmd}
				return command.Action(ctx, wrapped)
			},
		})
	}
	return commands
}

type CommandContext interface {
	String(name string) string
	Int(name string) int
	Bool(name string) bool
	Args() []string
}

type UrfaveContext struct {
	cmd *cli.Command
}

func (u *UrfaveContext) String(name string) string {
	return u.cmd.String(name)
}

func (u *UrfaveContext) Int(name string) int {
	return u.cmd.Int(name)
}

func (u *UrfaveContext) Bool(name string) bool {
	return u.cmd.Bool(name)
}

func (u *UrfaveContext) Args() []string {
	return u.cmd.Args().Slice()
}
