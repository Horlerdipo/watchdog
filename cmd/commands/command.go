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
		var arguments []cli.Argument
		//arguments := []cli.Argument{
		//	&cli.StringArg{
		//		Name:      "url",
		//		UsageText: "The URL of the watchdog monitoring process.",
		//	},
		//	&cli.StringArg{
		//		Name:      "http_method",
		//		Value:     "get",
		//		UsageText: "The HTTP method that URL will be called with. Options are: get,post,patch,put,delete",
		//	},
		//	&cli.StringArg{
		//		Name:      "frequency",
		//		Value:     "five_minutes",
		//		UsageText: "The Frequency the URL will be called. Options are: ten_seconds,thirty_seconds,five_minutes,thirty_minutes,one_hour,twelve_hours,twenty_four_hours",
		//	},
		//	&cli.StringArg{
		//		Name:      "contact_email",
		//		UsageText: "The email an alert will be sent to if the URL is unreachable.",
		//	},
		//}

		for _, argument := range command.Arguments() {
			var transformedArgument cli.Argument
			if argument.Type == enums.Int {
				transformedArgument = &cli.IntArg{
					Name:      argument.Name,
					UsageText: "The email an alert will be sent to if the URL is unreachable.",
				}
			} else {
				transformedArgument = &cli.StringArg{
					Name:      argument.Name,
					UsageText: "The email an alert will be sent to if the URL is unreachable.",
				}
			}
			arguments = append(arguments, transformedArgument)
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

type ArgumentContext struct {
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

func (u *UrfaveContext) Bool(name string) bool {
	return u.cmd.Bool(name)
}

func (u *UrfaveContext) Args() []string {
	return u.cmd.Args().Slice()
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
