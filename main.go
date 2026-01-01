package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jabreu610/gator/internal/config"
	"github.com/jabreu610/gator/internal/database"
	"github.com/jabreu610/gator/internal/rss"
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

func getCurrentUser(ctx context.Context, s *state) (database.User, error) {
	u, err := s.db.GetUserByName(ctx, s.config.CurrentUser)
	if err != nil {
		return database.User{}, err
	}
	return u, nil
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

func handlerUsers(s *state, cmd command) error {
	currentUser := s.config.CurrentUser
	u, err := s.db.ListUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range u {
		entry := fmt.Sprintf("* %s", user.Name)
		if user.Name == currentUser {
			entry += " (current)"
		}
		fmt.Println(entry)
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	var feedURL string
	if len(cmd.args) < 1 {
		// return errors.New("'agg' command expects one argument, the feed url")
		feedURL = "https://www.wagslane.dev/index.xml"
	} else {
		feedURL = cmd.args[0]
	}
	r, err := rss.FetchFeed(context.Background(), feedURL)
	if err != nil {
		return err
	}
	p, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fmt.Printf("%+v\n", r)
		return nil
	}
	fmt.Println(string(p))
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("'AddFeed' command expects two arguments: the name of the feed and the url")
	}
	context := context.Background()
	u, err := getCurrentUser(context, s)
	if err != nil {
		return err
	}
	newFeed := database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Name:   cmd.args[0],
		Url:    cmd.args[1],
		UserID: u.ID,
	}
	f, err := s.db.CreateFeed(context, newFeed)
	p, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		fmt.Printf("%+v\n", f)
		return nil
	}
	fmt.Println(string(p))
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	f, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return err
	}
	p, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		fmt.Printf("%+v\n", f)
		return nil
	}
	fmt.Println(string(p))
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
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerFeeds)

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
