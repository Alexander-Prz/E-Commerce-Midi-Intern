package cli

import (
	"GamesProject/internal/auth"
	"GamesProject/internal/services"
	"GamesProject/internal/utils"
	"context"
	"fmt"
	"time"
)

func Adm_Menu() {
	for {
		fmt.Println("\n=== ADMIN MODE ===")
		fmt.Println("[1] Game Catalog")
		fmt.Println("[2] Add New Genre")
		fmt.Println("[3] Remove Genre")
		fmt.Println("[4] Add Developer Account")
		fmt.Println("[5] Transaction Report")
		fmt.Println("[6] User List")
		fmt.Println("[7] Developer List")
		fmt.Println("[0] Logout")

		choice := utils.ReadChoice("=> ", 0, 387)
		switch choice {
		case 1:
			utils.ClearTerminal()
			Adm_GameCatalog()
		case 2:
			utils.ClearTerminal()
			Adm_AddGenre()
		case 3:
			utils.ClearTerminal()
			Adm_RemoveGenre()
		case 4:
			utils.ClearTerminal()
			Adm_AddDev()
		case 5:
			utils.ClearTerminal()
			Adm_Transaction()
		case 6:
			utils.ClearTerminal()
			Adm_UserList()
		case 7:
			utils.ClearTerminal()
			Adm_DeveloperList()
		case 387:
			utils.ClearTerminal()
			Adm_AllAccounts()
		case 0:
			if !utils.ReadConfirmation("Are you sure you want to logout? (y/n): ") {
				utils.ClearTerminal()
				continue
			}
			auth.Logout()
			utils.ClearTerminal()
			return
		default:
			fmt.Println("Please input a valid choice!")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
		}
	}
}

func Adm_GameCatalog() {
	ctx := context.Background()
	page := 1

	for {
		games, totalPages, err := services.AllGames(ctx, page)
		if err != nil {
			fmt.Println("Error loading games:", err)
			return
		}

		// Auto-adjust page if out-of-range (example: after deleting items)
		if page > totalPages && totalPages > 0 {
			page = totalPages
			continue
		}

		fmt.Println("\n=== GAME CATALOG ===")

		if totalPages == 0 {
			fmt.Println("No games found.")
			return
		}

		for i, g := range games {
			fmt.Printf("[%d] %s | Game ID: %d\n", i+1, g.Title, g.GameID)
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter Game ID to view, or 0 to go back")

		input := utils.ReadPagingInput("=> ")

		// Back
		if input.Command == "" && input.ID == 0 {
			return // go back
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

		// ID selection
		if input.ID > 0 {
			utils.ClearTerminal()
			Adm_GameMenu(input.ID)
			continue
		}

		fmt.Println("Invalid input.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
	}
}

func Adm_GameMenu(gameID int) {
	for {
		services.GameDetails(gameID)

		fmt.Println("\n=== GAME OPTIONS ===")
		fmt.Println("[1] Remove Game")
		fmt.Println("[0] Back")

		choice := utils.ReadChoice("=> ", 0, 1)

		switch choice {
		case 1:
			removed := Adm_RemoveGame(gameID)
			if removed {
				// Exit this menu so the catalog reloads
				return
			}
		case 0:
			utils.ClearTerminal()
			return
		default:
			fmt.Println("Invalid choice.")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
		}
	}
}

func Adm_RemoveGame(gameID int) bool {
	ctx := context.Background()

	if !utils.ReadConfirmation("Are you sure you want to remove this game? (y/n): ") {
		fmt.Println("Cancelled.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return false
	}

	err := services.RemoveGame(ctx, gameID, auth.CurrentUser.Role, 0)
	if err != nil {
		fmt.Println("Failed to remove game:", err)
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return false
	}

	fmt.Println("Game removed successfully.")
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
	return true
}

func Adm_AddGenre() {
	ctx := context.Background()

	name := utils.ReadLine("Genre Name: ")

	err := services.AddGenre(ctx, name)
	if err != nil {
		fmt.Println("Failed to add genre:", err)
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return
	}

	fmt.Println("Genre added successfully.")
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
}

func Adm_RemoveGenre() {
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

		fmt.Println("\n=== GENRE LIST ===")
		for _, g := range genres {
			fmt.Printf("[%d] %s\n", g.GenreID, g.GenreName)
		}

		if totalPages == 0 {
			fmt.Println("No genres found.")
			return
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter Genre ID to remove, or 0 to go back")

		input := utils.ReadPagingInput("=> ")

		// Back
		if input.Command == "" && input.ID == 0 {
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

		// Remove ID
		if input.ID > 0 {
			if !utils.ReadConfirmation("Are you sure you want to remove this genre? (y/n): ") {
				fmt.Println("Cancelled.")
				utils.ClearTerminal()
				continue
			}

			err := services.RemoveGenre(ctx, input.ID)
			if err != nil {
				fmt.Println("Failed to remove genre:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			} else {
				fmt.Println("Genre removed successfully.")
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

func Adm_AddDev() {
	ctx := context.Background()

	fmt.Println("\n=== ADD NEW DEVELOPER ===")

	email := utils.ReadEmail("Email: ")

	password, err := utils.ReadPasswordMasked("Password: ")
	if err != nil {
		fmt.Println("Error reading password:", err)
		return
	}

	devName := utils.ReadLine("Developer Name: ")

	if !utils.ReadConfirmation("Create this developer account? (y/n): ") {
		fmt.Println("Cancelled.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return
	}

	err = auth.RegisterForDeveloper(ctx, email, password, devName)
	if err != nil {
		fmt.Println("Failed to create developer:", err)
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return
	}

	fmt.Println("Developer account created successfully!")
	time.Sleep(1000 * time.Millisecond)
	utils.ClearTerminal()
}

func Adm_Transaction() {
	ctx := context.Background()

	list, err := services.GetAllTransactions(ctx)
	if err != nil {
		fmt.Println("Failed to load transactions:", err)
		return
	}

	if len(list) == 0 {
		fmt.Println("No transactions found.")
		fmt.Println()
		utils.ReadChoice("0. Back => ", 0, 0)
		return
	}

	fmt.Println("\n=== TRANSACTION REPORT ===")
	fmt.Println("|   Order ID   | Customer |  Total |       Date       |")
	for i, t := range list {
		fmt.Printf("| [%d] Order #%d |  Cust %d  | $%.2f | %s |\n",
			i+1, t.OrderID, t.CustomerID, t.TotalPrice,
			t.OrderDate.Format("2006-01-02 15:04"),
		)
	}

	fmt.Println("[0] Back")
	utils.ReadChoice("=> ", 0, 0)
}

func Adm_UserList() {
	ctx := context.Background()
	page := 1
	pageSize := 10 // adjust page size as needed

	for {
		users, err := services.GetAllUsers(ctx)
		if err != nil {
			fmt.Println("Failed to load users:", err)
			return
		}

		totalUsers := len(users)
		totalPages := totalUsers / pageSize
		if totalUsers%pageSize != 0 {
			totalPages++
		}

		// Fix page overflow
		if page > totalPages && totalPages > 0 {
			page = totalPages
		}

		start := (page - 1) * pageSize
		end := start + pageSize
		if end > totalUsers {
			end = totalUsers
		}

		fmt.Println("\n=== USER LIST ===")
		for i, u := range users[start:end] {
			status := ""
			if u.DeletedAt != nil {
				status = " (BANNED)"
			}
			fmt.Printf("[%d] %s | Role: %s | Auth ID: %d %s\n", i+1, u.Email, u.Role, u.AuthID, status)
		}

		if totalPages == 0 {
			fmt.Println("No users found.")
			return
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter User ID to view, or 0 to go back")

		input := utils.ReadPagingInput("=> ")

		// Exit
		if input.Command == "" && input.ID == 0 {
			utils.ClearTerminal()
			return
		}

		// Pagination
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

		// View user detail
		if input.ID > 0 {
			utils.ClearTerminal()
			Adm_AccountDetail(input.ID)
			continue
		}

		fmt.Println("Invalid input.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
	}
}

func Adm_DeveloperList() {
	ctx := context.Background()
	page := 1
	pageSize := 10

	for {
		devs, err := services.GetAllDevelopers(ctx)
		if err != nil {
			fmt.Println("Failed to load developers:", err)
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			return
		}

		totalDevs := len(devs)
		totalPages := totalDevs / pageSize
		if totalDevs%pageSize != 0 {
			totalPages++
		}

		// Fix page overflow
		if page > totalPages && totalPages > 0 {
			page = totalPages
		}

		start := (page - 1) * pageSize
		end := start + pageSize
		if end > totalDevs {
			end = totalDevs
		}

		fmt.Println("\n=== DEVELOPER LIST ===")
		for i, d := range devs[start:end] {
			status := ""
			if d.DeletedAt != nil || d.AuthDeleted != nil {
				status = " (BANNED)"
			}
			fmt.Printf("[%d] %s | Auth ID: %d%s\n",
				i+1, d.DeveloperName, d.AuthID, status)
		}

		if totalPages == 0 {
			fmt.Println("No developers found.")
			return
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter Developer AuthID to view, or 0 to go back")

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

		// View account detail by AuthID
		if input.ID > 0 {
			Adm_AccountDetail(input.ID)
			utils.ClearTerminal()
			continue
		}

		fmt.Println("Invalid input.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
	}
}

func Adm_AllAccounts() {
	ctx := context.Background()
	page := 1
	pageSize := 10 // adjust as needed

	for {
		// Fetch all accounts
		all, err := services.GetAllAccounts(ctx)
		if err != nil {
			fmt.Println("Failed to load accounts:", err)
			return
		}

		totalAccounts := len(all)
		totalPages := totalAccounts / pageSize
		if totalAccounts%pageSize != 0 {
			totalPages++
		}

		// Fix page overflow
		if page > totalPages && totalPages > 0 {
			page = totalPages
		}

		start := (page - 1) * pageSize
		end := start + pageSize
		if end > totalAccounts {
			end = totalAccounts
		}

		fmt.Println("\n=== ALL ACCOUNTS ===")
		for i, a := range all[start:end] {
			status := ""
			if a.DeletedAt != nil || (a.Role == "developer" && a.AuthDeleted != nil) {
				status = " (BANNED)"
			}

			if a.Role == "developer" && a.DeveloperName != nil {
				fmt.Printf("[%d] %s | Role: %s | Dev: %s | Auth ID: %d%s\n",
					i+1, a.Email, a.Role, *a.DeveloperName, a.AuthID, status)
			} else {
				fmt.Printf("[%d] %s | Role: %s | Auth ID: %d%s\n",
					i+1, a.Email, a.Role, a.AuthID, status)
			}
		}

		if totalPages == 0 {
			fmt.Println("No accounts found.")
			return
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter AuthID to view, or 0 to go back")

		input := utils.ReadPagingInput("=> ")

		// Exit
		if input.Command == "" && input.ID == 0 {
			utils.ClearTerminal()
			return
		}

		// Pagination
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

		// View account detail
		if input.ID > 0 {
			utils.ClearTerminal()
			Adm_AccountDetail(input.ID)
			continue
		}

		fmt.Println("Invalid input.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
	}
}

func Adm_AccountDetail(authID int) {
	ctx := context.Background()

	for {
		a, err := services.GetAccountByAuthID(ctx, authID)
		if err != nil {
			fmt.Println("Account not found:", err)
			return
		}

		status := "Active"
		if a.DeletedAt != nil || (a.Role == "developer" && a.AuthDeleted != nil) {
			status = "BANNED"
		}

		fmt.Println("\n=== ACCOUNT DETAIL ===")
		fmt.Printf("Auth ID   : %d\n", a.AuthID)
		fmt.Printf("Email     : %s\n", a.Email)
		fmt.Printf("Role      : %s\n", a.Role)
		fmt.Printf("Created At: %s\n", a.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Status    : %s\n", status)

		if a.Role == "developer" && a.DeveloperName != nil {
			fmt.Printf("Developer ID   : %d\n", *a.DeveloperID)
			fmt.Printf("Developer Name : %s\n", *a.DeveloperName)
			fmt.Printf("Auth Deleted   : %v\n", a.AuthDeleted != nil)
		}

		fmt.Println("\n=== OPTIONS ===")

		switch a.Role {
		case "user":
			if a.DeletedAt == nil {
				fmt.Println("[1] Ban User")
			} else {
				fmt.Println("[1] Unban User")
			}
		case "developer":
			fmt.Println("[1] (Cannot ban developer account)")
		default:
			fmt.Println("[1] (Cannot modify this account)")
		}

		fmt.Println("[0] Back")
		choice := utils.ReadChoice("=> ", 0, 1)

		if choice == 0 {
			utils.ClearTerminal()
			return
		}

		if a.Role != "user" {
			fmt.Println("Cannot modify this account.")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			continue
		}

		if a.DeletedAt == nil {
			if utils.ReadConfirmation("Ban this user? (y/n): ") {
				err := services.BanUser(ctx, a.AuthID)
				if err != nil {
					fmt.Println("Failed to ban user:", err)
					time.Sleep(1000 * time.Millisecond)
					utils.ClearTerminal()
					continue
				}
				fmt.Println("User banned successfully.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			} else {
				fmt.Println("Cancelled")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}
		} else {
			if utils.ReadConfirmation("Unban this user? (y/n): ") {
				err := services.UnbanUser(ctx, a.AuthID)
				if err != nil {
					fmt.Println("Failed to unban user:", err)
					time.Sleep(1000 * time.Millisecond)
					utils.ClearTerminal()
					continue
				}
				fmt.Println("User unbanned successfully.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			} else {
				fmt.Println("Cancelled")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}
		}
	}
}
