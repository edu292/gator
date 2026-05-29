package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gator/internal/config"
	"gator/internal/database"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	name string
	args []string
}

func handlerLogin(s *state, cmd command) error {
	if numArgs := len(cmd.args); numArgs != 1 {
		return fmt.Errorf("invalid number of args. expected: 1. got: %d", numArgs)
	}

	userName := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		log.Fatalf("User %s does not exist", userName)
	}

	s.cfg.SetUser(userName)
	fmt.Printf("Current user has been set to: %s\n", userName)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if numArgs := len(cmd.args); numArgs != 1 {
		return fmt.Errorf("invalid number of args. expected: 1. got: %d", numArgs)
	}

	userName := cmd.args[0]
	_, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		Name:      userName,
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Fatalf("User %s already is registered", userName)
	}

	s.cfg.SetUser(userName)
	fmt.Printf("User %s has been created\n", userName)
	return nil
}

type commands struct {
	registry map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, exist := c.registry[cmd.name]
	if !exist {
		log.Fatalf("Command %s does not exist", cmd.name)
	}

	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registry[name] = f
}

func main() {
	c, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", c.DBUrl)
	if err != nil {
		log.Fatalf("Could not connect to db: %s", err)
	}

	dbQueries := database.New(db)

	s := &state{cfg: c, db: dbQueries}
	cmds := commands{make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	args := os.Args
	if len(args) <= 2 {
		log.Fatalf("Too few args")
	}

	cmd := command{name: args[1], args: args[2:]}
	cmds.run(s, cmd)
}
