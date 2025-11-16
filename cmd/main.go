package main

import (
	"context"
	"github.com/horlerdipo/watchdog/cmd/commands"
	"github.com/urfave/cli/v3"
	"log"
	"os"
)

var command commands.CommandContainer

func main() {
	ctx := context.Background()
	ctx = context.WithoutCancel(ctx)

	cmd := &cli.Command{
		Name:     "Watchdog",
		Usage:    "A HTTP monitoring service written in Golang",
		Flags:    []cli.Flag{},
		Commands: command.Initiate(),
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
