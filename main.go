package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jabreu610/gator/internal/config"
	"github.com/jabreu610/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	config *config.Config
	db     *database.Queries
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
	u, err := s.db.GetUserByName(context.Background(), cmd.args[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user %s does not exist, register the user first", cmd.args[0])
		}
		return err
	}

	if err := s.config.SetUser(u.Name); err != nil {
		return err
	}
	fmt.Printf("'%s' now logged in\n", u.Name)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("'login' command expects one argument, the username")
	}
	name := cmd.args[0]
	newUser := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Name: name,
	}
	u, err := s.db.CreateUser(context.Background(), newUser)
	if err != nil {
		return err
	}
	if err := s.config.SetUser(u.Name); err != nil {
		return err
	}
	fmt.Printf("User successfully registered and logged in: %+v", u)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.ClearUsers(context.Background()); err != nil {
		return fmt.Errorf("Failed to reset users table: %w", err)
	}
	fmt.Println("Successfully cleared user table")
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
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)

	db, err := sql.Open("postgres", c.DBURL)
	if err != nil {
		fmt.Println("Unable to establish connection to database, please verify database connection string in configuration.")
		os.Exit(1)
	}
	dbQueries := database.New(db)
	s.db = dbQueries

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
