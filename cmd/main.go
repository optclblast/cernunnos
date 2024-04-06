package main

import (
	"cernunnos/internal/pkg/config"
	"cernunnos/internal/pkg/logger"
	"cernunnos/internal/server"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "cernunnos",
		Version: "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log-level",
				Value: "debug",
			},
			&cli.StringFlag{
				Name:  "address",
				Value: "localhost:8080",
			},
			&cli.StringFlag{
				Name: "db-host",
			},
			&cli.StringFlag{
				Name: "db-user",
			},
			&cli.StringFlag{
				Name: "db-password",
			},
		},
		Action: func(c *cli.Context) error {
			log := logger.NewLogger(logger.MapLevel(c.String("log-level")))
			log.Debug(
				"flags passed",
				slog.String("log-level", c.String("log-level")),
				slog.String("address", c.String("address")),
				slog.String("db-host", c.String("db-host")),
				slog.String("db-user", c.String("db-user")),
				slog.String("db-password", c.String("db-password")),
			)

			cfg := config.Config{
				Address:          c.String("address"),
				LogLevel:         c.String("log-level"),
				DatabaseHost:     c.String("db-host"),
				DatabaseUser:     c.String("db-user"),
				DatabasePassword: c.String("db-password"),
			}

			server, cleanup, err := server.ProvideServer(&cfg)
			if err != nil {
				return fmt.Errorf("error initialize server. %w", err)
			}

			defer cleanup()

			server.Start()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
