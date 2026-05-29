package commands

import (
	"context"
	"fmt"
	"log"
	"time"

	"gator/internal/config"
	"gator/internal/database"
	"gator/internal/rss"

	"github.com/google/uuid"
)

type command struct {
	name string
	args []string
}

func NewCommand(args []string) command {
	return command{name: args[1], args: args[2:]}
}

func handlerLogin(s *config.State, cmd command) error {
	if numArgs := len(cmd.args); numArgs != 1 {
		return fmt.Errorf("invalid number of args. expected: 1. got: %d", numArgs)
	}

	userName := cmd.args[0]
	_, err := s.DB.GetUser(context.Background(), userName)
	if err != nil {
		log.Fatalf("User %s does not exist", userName)
	}

	s.Cfg.SetUser(userName)
	fmt.Printf("Current user has been set to: %s\n", userName)
	return nil
}

func handlerRegister(s *config.State, cmd command) error {
	if numArgs := len(cmd.args); numArgs != 1 {
		return fmt.Errorf("invalid number of args. expected: 1. got: %d", numArgs)
	}

	userName := cmd.args[0]
	_, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		Name:      userName,
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Fatalf("User %s already is registered", userName)
	}

	s.Cfg.SetUser(userName)
	fmt.Printf("User %s has been created\n", userName)
	return nil
}

func handlerReset(s *config.State, cmd command) error {
	err := s.DB.ResetUsers(context.Background())
	if err != nil {
		log.Fatalf("Failed to reset users: %s", err)
	}

	fmt.Println("Reset successfull")
	return nil
}

func handlerUsers(s *config.State, cmd command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		log.Fatalf("Could not retrive users from DB: %s", err)
	}

	for _, user := range users {
		fmt.Print("* ", user.Name)
		if user.Name == s.Cfg.CurrentUserName {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}

	return nil
}

func handlerAgg(s *config.State, cmd command) error {
	feed, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*feed)
	return nil
}

func handlerAddFeed(s *config.State, cmd command) error {
	user
}

type commands struct {
	registry map[string]func(*config.State, command) error
}

func NewCommands() commands {
	cmds := commands{make(map[string]func(*config.State, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)

	return cmds
}

func (c *commands) Run(s *config.State, cmd command) error {
	handler, exist := c.registry[cmd.name]
	if !exist {
		log.Fatalf("Command %s does not exist", cmd.name)
	}

	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*config.State, command) error) {
	c.registry[name] = f
}
