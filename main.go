package main

import (
	"log"
	"os"

	"1001-twacc-chat/internal/app"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "seed-admin" {
		if err := app.SeedAdmin(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
