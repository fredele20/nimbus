package main

import (
	"log/slog"
	"os"

	"github.com/fredele20/nimbus/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		slog.Error("application terminated", "error", err)
		os.Exit(1)
	}
}
