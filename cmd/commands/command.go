package commands

import (
	"context"
	"github.com/horlerdipo/watchdog/database"
	"github.com/horlerdipo/watchdog/enums"
	"github.com/horlerdipo/watchdog/orchestrator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v3"
)

type Command interface {
	Name() string
	Action(ctx context.Context, cmd CommandContext) error
	Aliases() []string
	Usage() string
	Arguments() []ArgumentContext
	Flags() []FlagContext
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
	cc.Register(NewRemoveCommand())
	cc.Register(NewListCommand())
}

func (cc *CommandContainer) Initiate() []*cli.Command {
	cc.RegisterAll()
	var commands []*cli.Command
	for _, command := range cc.Commands {
		var arguments []cli.Argument
		var flags []cli.Flag

		for _, argument := range command.Arguments() {
			var transformedArgument cli.Argument
			if argument.Type == enums.Int {
				transformedArgument = &cli.IntArg{
					Name:      argument.Name,
					UsageText: argument.Usage,
				}
			} else {
				transformedArgument = &cli.StringArg{
					Name:      argument.Name,
					UsageText: argument.Usage,
				}
			}
			arguments = append(arguments, transformedArgument)
		}

		for _, flag := range command.Flags() {
			var transformedFlag cli.Flag
			if flag.Type == enums.Int {
				transformedFlag = &cli.IntFlag{
					Name:  flag.Name,
					Usage: flag.Usage,
					Value: flag.Default.(int),
				}
			} else {
				transformedFlag = &cli.StringFlag{
					Name:  flag.Name,
					Usage: flag.Usage,
					Value: flag.Default.(string),
				}
			}
			flags = append(flags, transformedFlag)
		}

		commands = append(commands, &cli.Command{
			Name:    command.Name(),
			Usage:   command.Usage(),
			Aliases: command.Aliases(),
			Action: func(ctx context.Context, cmd *cli.Command) error {
				wrapped := &UrfaveContext{cmd: cmd}
				return command.Action(ctx, wrapped)
			},
			Arguments: arguments,
			Flags:     flags,
		})
	}
	return commands
}

type CommandContext interface {
	String(name string) string
	Int(name string) int
	Args() []string
	BoolFlag(name string) bool
	IntFlag(name string) int
	StringFlag(name string) string
}

type ArgumentContext struct {
	Name    string
	Usage   string
	Type    enums.ArgumentType
	Default interface{}
}

type FlagContext struct {
	Name    string
	Usage   string
	Type    enums.ArgumentType
	Default interface{}
}
type UrfaveContext struct {
	cmd *cli.Command
}

func (u *UrfaveContext) String(name string) string {
	return u.cmd.StringArg(name)
}

func (u *UrfaveContext) Int(name string) int {
	return u.cmd.IntArg(name)
}

func (u *UrfaveContext) Args() []string {
	return u.cmd.Args().Slice()
}

func (u *UrfaveContext) BoolFlag(name string) bool {
	return u.cmd.Bool(name)
}

func (u *UrfaveContext) StringFlag(name string) string {
	return u.cmd.String(name)
}

func (u *UrfaveContext) IntFlag(name string) int {
	return u.cmd.Int(name)
}

func RefreshRedisInterval(ctx context.Context, redisClient *redis.Client, pool *pgxpool.Pool, frequency enums.MonitoringFrequency) error {
	urls, err := database.NewUrlRepository(pool).FetchAll(ctx, 10, 0, database.UrlQueryFilter{
		Frequency: frequency,
	})
	if err != nil {
		return err
	}

	redisClient.Del(ctx, orchestrator.FormatRedisList(frequency.ToSeconds()))

	for _, url := range urls {
		redisClient.LPush(ctx, orchestrator.FormatRedisList(url.MonitoringFrequency.ToSeconds()), url.Url)
	}
	return nil
}
