package cli

import (
	"GamesProject/internal/auth"
	"GamesProject/internal/services"
	"GamesProject/internal/utils"
	"context"
	"fmt"
	"time"
)

func Dev_Menu() {
	ctx := context.Background()

	if auth.CurrentUser == nil {
		fmt.Println("No user logged in.")
		return
	}

	devID, _, err := services.GetDeveloperByAuthID(ctx, auth.CurrentUser.AuthID)
	if err != nil {
		fmt.Println("Cannot find developer profile:", err)
		return
	}

	for {
		fmt.Println("\n=== DEVELOPER MENU ===")
		fmt.Println("[1] View My Games")
		fmt.Println("[2] Add Game")
		fmt.Println("[3] View Sales Report")
		fmt.Println("[0] Logout")

		choice := utils.ReadChoice("=> ", 0, 3)

		switch choice {
		case 1:
			utils.ClearTerminal()
			Dev_GameCatalog(devID)
		case 2:
			utils.ClearTerminal()
			Dev_AddGame(devID)
		case 3:
			utils.ClearTerminal()
			Dev_Sales(devID)
		case 0:
			if !utils.ReadConfirmation("Are you sure you want to logout? (y/n): ") {
				utils.ClearTerminal()
				continue
			}
			utils.ClearTerminal()
			auth.Logout()
			return
		default:
			fmt.Println("Invalid choice.")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
		}
	}
}

func Dev_GameCatalog(devID int) {
	ctx := context.Background()
	page := 1

	for {
		games, totalPages, err := services.DeveloperGames(ctx, devID, page)
		if err != nil {
			fmt.Println("Error loading games:", err)
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			return
		}

		if page > totalPages && totalPages > 0 {
			page = totalPages
			continue
		}

		fmt.Println("\n=== MY GAMES ===")
		if totalPages == 0 {
			fmt.Println("No games found.")
			return
		}

		for i, g := range games {
			fmt.Printf("[%d] %s | Game ID: %d\n", i+1, g.Title, g.GameID)
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter Game ID to manage, or 0 to go back")

		input := utils.ReadPagingInput("=> ")

		if input.Command == "" && input.ID == 0 {
			utils.ClearTerminal()
			return
		}

		if input.Command == "<" {
			if page > 1 {
				page--
				utils.ClearTerminal()
			} else {
				fmt.Println("Already at first page.")
				utils.ClearTerminal()
			}
			continue
		}
		if input.Command == ">" {
			if page < totalPages {
				page++
				utils.ClearTerminal()
			} else {
				fmt.Println("Already at last page.")
				utils.ClearTerminal()
			}
			continue
		}

		if input.ID > 0 {
			owned, err := services.GameOwnedByDeveloper(ctx, devID, input.ID)
			if err != nil {
				fmt.Println("Error checking permission:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}

			if !owned {
				fmt.Println("You do not have permission to manage this game.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}

			utils.ClearTerminal()
			Dev_ManageGameMenu(devID, input.ID)
			continue
		}

		fmt.Println("Invalid input.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
	}
}

func Dev_ManageGameMenu(devID, gameID int) {
	ctx := context.Background()
	for {
		services.GameDetails(gameID)
		fmt.Printf("\n=== MANAGE GAME %d ===\n", gameID)
		fmt.Println("[1] Edit")
		fmt.Println("[2] Remove")
		fmt.Println("[3] Add Genre")
		fmt.Println("[4] Edit Genres")
		fmt.Println("[0] Back")

		choice := utils.ReadChoice("=> ", 0, 4)
		switch choice {
		case 1:
			if err := Dev_EditGameByID(devID, gameID); err != nil {
				fmt.Println("Edit failed:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			}
		case 2:
			if !utils.ReadConfirmation("Are you sure you want to remove this game? (y/n): ") {
				fmt.Println("Cancelled.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}
			if err := services.RemoveGame(ctx, gameID, auth.CurrentUser.Role, devID); err != nil {
				fmt.Println("Failed to remove game:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			} else {
				fmt.Println("Game removed.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			}
			return
		case 3:
			utils.ClearTerminal()
			Dev_AddGameGenre(gameID)
		case 4:
			utils.ClearTerminal()
			Dev_EditGameGenre(gameID)
		case 0:
			utils.ClearTerminal()
			return
		default:
			fmt.Println("Invalid.")
		}
	}
}

func Dev_AddGameGenre(gameID int) {
	ctx := context.Background()
	page := 1

	for {
		genres, totalPages, err := services.AllGenres(ctx, page)
		if err != nil {
			fmt.Println("Error loading genres:", err)
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			return
		}

		if page > totalPages && totalPages > 0 {
			page = totalPages
			continue
		}

		fmt.Printf("\n=== ADD GENRE TO GAME %d ===\n", gameID)
		for _, g := range genres {
			fmt.Printf("[%d] %s\n", g.GenreID, g.GenreName)
		}

		if totalPages == 0 {
			fmt.Println("No genres available.")
			return
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter Genre ID to add, or 0 to go back")

		input := utils.ReadPagingInput("=> ")

		// Back
		if input.Command == "" && input.ID == 0 {
			utils.ClearTerminal()
			return
		}

		// Prev page
		if input.Command == "<" {
			if page > 1 {
				page--
				utils.ClearTerminal()
			} else {
				fmt.Println("Already at first page.")
				utils.ClearTerminal()
			}
			continue
		}

		// Next page
		if input.Command == ">" {
			if page < totalPages {
				page++
				utils.ClearTerminal()
			} else {
				fmt.Println("Already at last page.")
				utils.ClearTerminal()
			}
			continue
		}

		// Add genre
		if input.ID > 0 {
			err := services.AddGenreToGame(ctx, gameID, input.ID)
			if err != nil {
				fmt.Println("Failed to add genre:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			} else {
				fmt.Println("Genre added successfully.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			}
			continue
		}

		fmt.Println("Invalid input.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
	}
}

func Dev_EditGameGenre(gameID int) {
	ctx := context.Background()
	page := 1

	for {
		genres, totalPages, err := services.AllGenres(ctx, page)
		if err != nil {
			fmt.Println("Error loading genres:", err)
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			return
		}

		if page > totalPages && totalPages > 0 {
			page = totalPages
			continue
		}

		fmt.Printf("\n=== EDIT GENRES FOR GAME %d ===\n", gameID)
		for _, g := range genres {
			fmt.Printf("[%d] %s\n", g.GenreID, g.GenreName)
		}

		if totalPages == 0 {
			fmt.Println("No genres available.")
			return
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter new Genre IDs (comma separated), or 0 to go back")

		raw := utils.ReadLine("=>")

		// Back
		if raw == "0" {
			utils.ClearTerminal()
			return
		}

		// Prev page
		if raw == "<" {
			if page > 1 {
				page--
				utils.ClearTerminal()
			} else {
				fmt.Println("Already at first page.")
				utils.ClearTerminal()
			}
			continue
		}

		// Next page
		if raw == ">" {
			if page < totalPages {
				page++
				utils.ClearTerminal()
			} else {
				fmt.Println("Already at last page.")
				utils.ClearTerminal()
			}
			continue
		}

		// Parse IDs
		genreIDs := utils.ParseIntList(raw)
		if len(genreIDs) == 0 {
			fmt.Println("Invalid input.")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			continue
		}

		if !utils.ReadConfirmation("Are you sure you want to replace all genres? (y/n): ") {
			fmt.Println("Cancelled.")
			utils.ClearTerminal()
			continue
		}

		// Update
		err = services.UpdateGameGenres(ctx, gameID, genreIDs)
		if err != nil {
			fmt.Println("Failed to update genres:", err)
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
		} else {
			fmt.Println("Genres updated successfully.")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
		}

		return
	}
}

func Dev_AddGame(devID int) {
	ctx := context.Background()
	title := utils.ReadLine("Title: ")
	price := utils.ReadFloat("Price: ")
	release := utils.ReadDate("Release Date (YYYY-MM-DD) or blank: ")

	id, err := services.AddGame(ctx, title, price, release, devID)
	if err != nil {
		fmt.Println("Failed to add game:", err)
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return
	}
	fmt.Println("Game added with ID:", id)
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
}

func Dev_EditGameByID(devID, gameID int) error {
	ctx := context.Background()
	title := utils.ReadLine("New Title: ")
	price := utils.ReadFloat("New Price: ")
	release := utils.ReadDate("New Release Date (YYYY-MM-DD) or blank: ")

	if err := services.EditGameDetails(ctx, gameID, title, price, release, devID); err != nil {
		return err
	}
	fmt.Println("Game updated.")
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
	return nil
}

func Dev_RemoveGame(devID int) {
	ctx := context.Background()

	id := utils.ReadInt("Game ID to remove: ")
	if !utils.ReadConfirmation("Are you sure? (y/n): ") {
		fmt.Println("Cancelled.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return
	}
	if err := services.RemoveGame(ctx, id, auth.CurrentUser.Role, devID); err != nil {
		fmt.Println("Failed to remove game:", err)
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return
	}
	fmt.Println("Game removed successfully.")
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
}

func Dev_Sales(devID int) {
	ctx := context.Background()

	for {
		fmt.Println("\n=== SALES REPORT ===")

		list, err := services.DeveloperSalesReport(ctx, devID)
		if err != nil {
			fmt.Println("Failed to load report:", err)
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			return
		}

		if len(list) == 0 {
			fmt.Println("You have no games or no sales.")
		} else {
			for _, r := range list {
				fmt.Printf("\nGame: %s (ID: %d)\n", r.Title, r.GameID)
				fmt.Printf("Units Sold: %d\n", r.UnitsSold)
				fmt.Printf("Revenue: $%.2f\n", r.Revenue)
			}
		}

		fmt.Println("\n[0] Back")

		input := utils.ReadChoice("=> ", 0, 0)

		if input == 0 {
			utils.ClearTerminal()
			return
		}

		fmt.Println("Invalid input.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
	}
}
