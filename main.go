package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"

	"gator/internal/commands"
	"gator/internal/config"
	"gator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", conf.DBUrl)
	if err != nil {
		log.Fatalf("Could not connect to db: %s", err)
	}

	dbQueries := database.New(db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)
	go func() {
		<-interruptChan
		cancel()
	}()

	s := config.NewState(conf, dbQueries)
	cmds := commands.NewRegistry()

	cmd, err := commands.NewCommand(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.Run(ctx, s, cmd)
	if err != nil {
		if errors.Is(err, commands.ErrFatal) {
			log.Fatal(err)
		}

		fmt.Println(err)
	}
}
