package main

import (
	"GamesProject/internal/cli"
	"GamesProject/internal/db"
	"GamesProject/internal/utils"
)

func main() {
	utils.ClearTerminal()

	utils.InitSignalHandler()

	pool, err := db.Connect()
	if err != nil {
		panic(err)
	}
	db.Pool = pool
	defer pool.Close()

	cli.ProgramStart()

}
