package main

import (
	"context"
	"github.com/horlerdipo/watchdog/cmd/commands"
	"github.com/horlerdipo/watchdog/env"
	"github.com/horlerdipo/watchdog/logger"
	"github.com/urfave/cli/v3"
	"log"
	"os"
)

var command commands.CommandContainer

func main() {
	ctx := context.Background()
	ctx = context.WithoutCancel(ctx)

	env.LoadEnv(".env")

	newLogger := logger.New()
	cmd := &cli.Command{
		Name:     "Watchdog",
		Usage:    "A HTTP monitoring service written in Golang",
		Flags:    []cli.Flag{},
		Commands: command.Initiate(newLogger),
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
