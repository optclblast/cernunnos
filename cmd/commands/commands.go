package commands

import "github.com/urfave/cli/v2"

var registry []*cli.Command

func Commands() []*cli.Command {
	return registry
}
func Register(command *cli.Command) {
	registry = append(registry, command)
}
