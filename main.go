package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jabreu610/gator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmdMap map[string]func(*state, command) error
}

func (c commands) run(s *state, cmd command) error {
	return c.cmdMap[cmd.name](s, cmd)
}

func (c commands) register(name string, f func(*state, command) error) {
	c.cmdMap[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("'login' command expects one argument, the username")
	}
	if err := s.config.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("'%s' now logged in\n", cmd.args[0])

	return nil
}

func main() {
	c, err := config.Read()
	if err != nil {
		msg := fmt.Errorf("Unable to read config file: %w", err)
		fmt.Println(msg)
	}

	s := state{
		config: &c,
	}
	commands := commands{
		cmdMap: make(map[string]func(*state, command) error),
	}
	commands.register("login", handlerLogin)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Expecting at least one argument")
		os.Exit(1)
	}
	commandArgs := args[1:]
	command := command{
		name: commandArgs[0],
		args: commandArgs[1:],
	}
	if err := commands.run(&s, command); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
