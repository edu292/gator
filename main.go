package main

import (
	"database/sql"
	"log"
	"os"

	"gator/internal/commands"
	"gator/internal/config"
	"gator/internal/database"

	_ "github.com/lib/pq"
)

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

	s := config.NewState(c, dbQueries)
	cmds := commands.NewCommands()
	args := os.Args
	if len(args) <= 1 {
		log.Fatalf("Too few args")
	}

	cmd := commands.NewCommand(args)
	cmds.Run(s, cmd)
}
