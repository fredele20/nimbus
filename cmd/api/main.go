package main

import (
	"log"

	"github.com/fredele20/nimbus/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
