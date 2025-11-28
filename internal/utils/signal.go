package utils

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"GamesProject/internal/db"

	"golang.org/x/term"
)

var originalState *term.State

func InitSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nInterrupted. Restoring terminal...")

		if originalState != nil {
			term.Restore(int(os.Stdin.Fd()), originalState)
		}

		// Close DB if needed
		if db.Pool != nil {
			db.Pool.Close()
		}

		os.Exit(1)
	}()
}
