package main

import (
	"log"

	"1001-twacc-chat/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
