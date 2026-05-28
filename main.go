package main

import (
	"fmt"
	"log"

	"gator/internal/config"
)

func main() {
	c, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	err = c.SetUser("edusk")
	if err != nil {
		log.Fatal(err)
	}

	c, _ = config.Read()
	fmt.Println(*c)
}
