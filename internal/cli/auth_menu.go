package cli

import (
	"GamesProject/internal/auth"
	"GamesProject/internal/utils"
	"context"
	"fmt"
	"time"
)

func ProgramStart() {
	ctx := context.Background()

	for {
		fmt.Println("MEONG!")
		fmt.Println("\n=== MEONG GAME SHOP ===")
		fmt.Println("[1] Login")
		fmt.Println("[2] Register")
		fmt.Println("[0] Exit")

		choice := utils.ReadChoice("=> ", 0, 387)

		switch choice {

		case 1:
			err := LoginUserInput(ctx)
			if err != nil {
				fmt.Println("Error:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}
			switch auth.CurrentUser.Role {
			case "admin":
				Adm_Menu()
			case "developer":
				Dev_Menu()
			default:
				User_Menu()
			}

		case 2:
			RegisterUserInput(ctx)

		case 387:
			RegisterAdminInput(ctx)

		case 0:
			if !utils.ReadConfirmation("Are you sure you want to quit? (y/n): ") {
				continue
			}
			fmt.Println("Thank you for visiting!")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			return
		}
	}
}

func RegisterUserInput(ctx context.Context) {

	email := utils.ReadEmail("Email: ")

	password, err := utils.ReadPasswordMasked("Password: ")
	if err != nil {
		fmt.Println("Error reading password:", err)
		return
	}

	username := utils.ReadLimitedWord("Username (max 20 character): ", 20)

	err = auth.Register(ctx, email, password, username)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Registration successful!")
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
}

func RegisterAdminInput(ctx context.Context) {

	email := utils.ReadEmail("Email: ")

	password, err := utils.ReadPasswordMasked("Password: ")
	if err != nil {
		fmt.Println("Error reading password:", err)
		return
	}

	err = auth.RegisterForAdmin(ctx, email, password)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Registration successful!")
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
}

func LoginUserInput(ctx context.Context) error {

	email := utils.ReadEmail("Email: ")

	password, err := utils.ReadPasswordMasked("Password: ")
	if err != nil {
		return fmt.Errorf("error reading password: %w", err)
	}

	err = auth.Login(ctx, email, password)
	if err != nil {
		return err
	}

	fmt.Println("Login successful! Welcome,", auth.CurrentUser.Username)
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
	return nil
}
