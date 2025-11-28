package cli

import (
	"GamesProject/internal/auth"
	"GamesProject/internal/services"
	"GamesProject/internal/utils"
	"context"
	"fmt"
	"sort"
	"time"
)

func User_Menu() {
	for {
		fmt.Printf("\n=== WELCOME, %s ===\n", auth.CurrentUser.Username)
		fmt.Println("[1] Game Catalog")
		fmt.Println("[2] Cart")
		fmt.Println("[3] Order History")
		fmt.Println("[0] Logout")

		choice := utils.ReadChoice("=> ", 0, 3)

		switch choice {
		case 1:
			utils.ClearTerminal()
			User_GameCatalog()
		case 2:
			utils.ClearTerminal()
			User_Cart()
		case 3:
			utils.ClearTerminal()
			User_OrderHistory()
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

func User_GameCatalog() {
	ctx := context.Background()
	page := 1

	for {
		games, totalPages, err := services.AllGames(ctx, page)
		if err != nil {
			fmt.Println("Error loading games:", err)
			return
		}

		// Fix page overflow
		if page > totalPages && totalPages > 0 {
			page = totalPages
			continue
		}

		fmt.Println("\n=== GAME CATALOG ===")
		for i, g := range games {
			fmt.Printf("[%d] %s | Game ID: %d\n", i+1, g.Title, g.GameID)
		}

		if totalPages == 0 {
			fmt.Println("No games found.")
			return
		}

		fmt.Printf("--- Page %d / %d ---\n", page, totalPages)
		fmt.Println("< Prev | Next >")
		fmt.Println("Enter Game ID to view, or 0 to go back")

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
			utils.ClearTerminal()
			User_GameMenu(input.ID)
			continue
		}

		fmt.Println("Invalid input.")
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
	}
}

func User_GameMenu(gameid int) {
	ctx := context.Background()
	services.GameDetails(gameid)

	fmt.Println("\n=== GAME OPTIONS ===")
	fmt.Println("[1] Add to Cart")
	fmt.Println("[0] Back")

	choice := utils.ReadChoice("=> ", 0, 1)
	switch choice {
	case 1:
		// === ADD TO CART LOGIC ===
		qty := utils.ReadInt("Quantity: ")

		gamePrice, err := services.GamePrice(ctx, gameid)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		err = services.AddToCart(ctx, auth.CurrentUser.CustomerID, gameid, qty, gamePrice)
		if err != nil {
			fmt.Println("Failed to add to cart:", err)
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
		} else {
			fmt.Println("Added to cart!")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
		}

	case 0:
		utils.ClearTerminal()
		return
	}
}

func User_Cart() {
	ctx := context.Background()

	for {
		cart, err := services.ViewCart(ctx, auth.CurrentUser.CustomerID)
		if err != nil {
			fmt.Println("Error loading cart:", err)
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()
			return
		}

		fmt.Println("\n=== YOUR CART ===")
		if len(cart.Items) == 0 {
			fmt.Println("Your cart is empty!")
			fmt.Println("[0] Back")
			utils.ReadChoice("=> ", 0, 0)
			utils.ClearTerminal()
			return
		}

		for i, item := range cart.Items {
			fmt.Printf("[%d] %s x%d (%.2f each) | Item ID: %d\n",
				i+1, item.Title, item.Quantity, item.PriceAtPurchase, item.OrderItemID)
		}

		fmt.Printf("Total: %.2f\n", cart.Total)

		fmt.Println("[1] Buy All Items")
		fmt.Println("[2] Remove Item")
		fmt.Println("[3] Clear Cart")
		fmt.Println("[0] Back")

		choice := utils.ReadChoice("=> ", 0, 3)
		switch choice {

		case 1:
			// === CHECKOUT & PAYMENT FLOW ===
			orderID, total, err := services.CheckoutCart(ctx, auth.CurrentUser.CustomerID)
			if err != nil {
				fmt.Println("Checkout failed:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}

			fmt.Println("\n=== PAYMENT METHODS ===")

			methods, err := services.ListPaymentMethods(ctx)
			if err != nil {
				fmt.Println("Failed to load payment methods:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}

			for _, m := range methods {
				fmt.Printf("[%d] %s\n", m.PaymentMethodID, m.Name)
			}

			methodChoice := utils.ReadInt("=> ")

			var chosenMethodID int
			for _, m := range methods {
				if m.PaymentMethodID == methodChoice {
					chosenMethodID = m.PaymentMethodID
					break
				}
			}

			if chosenMethodID == 0 {
				fmt.Println("Invalid payment method.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}

			pid, err := services.StartPaymentForOrder(ctx, orderID, chosenMethodID, total)
			if err != nil {
				fmt.Println("Failed to create payment:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}

			fmt.Println("Processing payment...")

			if err := services.ConfirmPayment(ctx, pid); err != nil {
				fmt.Println("Payment failed:", err)
				_ = services.FailPayment(ctx, pid)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
				continue
			}

			fmt.Println("Payment successful! Thank you.")
			time.Sleep(1000 * time.Millisecond)
			utils.ClearTerminal()

		case 2:
			id := utils.ReadInt("Enter OrderItemID to remove: ")

			err := services.RemoveFromCart(ctx, id)
			if err != nil {
				fmt.Println("Error:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			} else {
				fmt.Println("Item removed.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			}
			continue

		case 3:
			err := services.ClearCart(ctx, auth.CurrentUser.CustomerID)
			if err != nil {
				fmt.Println("Error:", err)
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			} else {
				fmt.Println("Cart cleared.")
				time.Sleep(1000 * time.Millisecond)
				utils.ClearTerminal()
			}
			continue

		case 0:
			return
		}
	}
}

func User_OrderHistory() {
	ctx := context.Background()

	history, err := services.GetOrderHistory(ctx, auth.CurrentUser.CustomerID)
	if err != nil {
		fmt.Println("Error loading order history:", err)
		time.Sleep(1000 * time.Millisecond)
		utils.ClearTerminal()
		return
	}

	fmt.Println("\n=== ORDER HISTORY ===")

	if len(history) == 0 {
		fmt.Println("No orders found.")
		fmt.Println("[0] Back")
		utils.ReadChoice("=> ", 0, 0)
		utils.ClearTerminal()
		return
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].OrderDate.Before(history[j].OrderDate)
	})

	for i, h := range history {
		fmt.Printf("Order #%d | $%.2f | %s\n",
			i+1,
			h.TotalPrice,
			h.OrderDate.Format("2006-01-02 15:04"),
		)

		fmt.Printf("Payment: %s", h.PaymentStatus)
		if h.PaidAt != nil {
			fmt.Printf(" at %s", h.PaidAt.Format("2006-01-02 15:04"))
		}
		fmt.Println()
		fmt.Println("---------------------------")
	}

	fmt.Println("[0] Back")
	utils.ReadChoice("=> ", 0, 0)
	utils.ClearTerminal()
}
