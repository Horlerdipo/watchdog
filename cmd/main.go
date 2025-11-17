package main

import (
	"context"
	"github.com/horlerdipo/watchdog/cmd/commands"
	"github.com/horlerdipo/watchdog/env"
	"github.com/urfave/cli/v3"
	"log"
	"os"
)

var command commands.CommandContainer

func main() {
	ctx := context.Background()
	ctx = context.WithoutCancel(ctx)

	env.LoadEnv(".env")

	//b, _ := json.MarshalIndent(command.Initiate(), "", "  ")
	//panic(string(b))

	cmd := &cli.Command{
		Name:     "Watchdog",
		Usage:    "A HTTP monitoring service written in Golang",
		Flags:    []cli.Flag{},
		Commands: command.Initiate(),
		//Commands: []*cli.Command{
		//	{
		//		Name:      "add",
		//		Aliases:   []string{"a"},
		//		Usage:     "Add a URL to the watchdog monitoring process.",
		//		ArgsUsage: "<url> [http_method] [frequency] [contact_email]",
		//		Arguments: []cli.Argument{
		//			&cli.StringArg{
		//				Name:      "url",
		//				UsageText: "The URL of the watchdog monitoring process.",
		//			},
		//			&cli.StringArg{
		//				Name:      "http_method",
		//				Value:     "get",
		//				UsageText: "The HTTP method that URL will be called with. Options are: get,post,patch,put,delete",
		//			},
		//			&cli.StringArg{
		//				Name:      "frequency",
		//				Value:     "five_minutes",
		//				UsageText: "The Frequency the URL will be called. Options are: ten_seconds,thirty_seconds,five_minutes,thirty_minutes,one_hour,twelve_hours,twenty_four_hours",
		//			},
		//			&cli.StringArg{
		//				Name:      "contact_email",
		//				UsageText: "The email an alert will be sent to if the URL is unreachable.",
		//			},
		//		},
		//		Action: func(ctx context.Context, cmd *cli.Command) error {
		//			// Fetch arguments by position
		//
		//			url := cmd.StringArg("url")
		//			httpMethod := cmd.StringArg("http_method")
		//			frequency := cmd.StringArg("frequency")
		//			contactEmail := cmd.StringArg("contact_email")
		//
		//			// Check if required argument is provided
		//			if url == "" {
		//				return fmt.Errorf("url is required")
		//			}
		//
		//			fmt.Printf("URL: %s\n", url)
		//			fmt.Printf("HTTP Method: %s\n", httpMethod)
		//			fmt.Printf("Frequency: %s\n", frequency)
		//			fmt.Printf("Contact Email: %s\n", contactEmail)
		//			return nil
		//		},
		//	},
		//},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
