package utils

import (
	"cernunnos/cmd/commands"
	"cernunnos/commands/utils"
	"cernunnos/internal/pkg/config"
	"cernunnos/internal/pkg/logger"
	"cernunnos/internal/usecase/repository"

	"github.com/urfave/cli/v2"
)

func init() {
	commands.Register(&cli.Command{
		Name: "fill-db",
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
			cfg := config.Config{
				Address:          c.String("address"),
				LogLevel:         c.String("log-level"),
				DatabaseHost:     c.String("db-host"),
				DatabaseUser:     c.String("db-user"),
				DatabasePassword: c.String("db-password"),
			}

			db, cleanup, err := repository.ProvideDatabaseConnection(&cfg)
			if err != nil {
				return err
			}
			defer cleanup()

			log := logger.NewLogger(logger.MapLevel(c.String("log-level")))

			command := utils.NewFillDatabaseCommand(db, log)

			if err := command.Run(c.Context); err != nil {
				panic(err)
			}
			return nil
		},
	})
}
