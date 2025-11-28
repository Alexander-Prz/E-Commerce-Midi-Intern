package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

var reader = bufio.NewReader(os.Stdin)

//
// ─── CLEAN INPUT ───────────────────────────────────────────────────────────────
//

// CleanInput removes leftover input (e.g., invalid characters)
func CleanInput() {
	reader.ReadString('\n')
}

//
// ─── READ LINE (SUPPORTS SPACES) ──────────────────────────────────────────────
//

// ReadLine reads a full line including spaces
func ReadLine(prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func ReadWord(prompt string) string {
	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read input. Try again.")
			continue
		}

		input = strings.TrimSpace(input)
		fields := strings.Fields(input)

		if len(fields) == 0 {
			fmt.Println("Input cannot be empty. Try again.")
			continue
		}

		return fields[0] // first word only, no spaces
	}
}

func ReadLimitedWord(prompt string, max int) string {
	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read input. Try again.")
			continue
		}

		input = strings.TrimSpace(input)
		fields := strings.Fields(input)
		if len(fields) == 0 {
			fmt.Println("Input cannot be empty.")
			continue
		}

		word := fields[0] // Only first word allowed

		if len(word) > max {
			fmt.Printf("Input too long! Maximum allowed is %d characters.\n", max)
			continue
		}

		return word
	}
}

//
// ─── READ INT SAFELY ───────────────────────────────────────────────────────────
//

// ReadInt keeps asking until the user enters a valid integer
func ReadInt(prompt string) int {
	for {
		fmt.Print(prompt)
		str, _ := reader.ReadString('\n')
		str = strings.TrimSpace(str)

		value, err := strconv.Atoi(str)
		if err == nil {
			return value
		}

		fmt.Println("Invalid number, please try again.")
	}
}

//
// ─── READ FLOAT SAFELY ─────────────────────────────────────────────────────────
//

// ReadFloat keeps asking until user enters a valid floating point number
func ReadFloat(prompt string) float64 {
	for {
		fmt.Print(prompt)
		str, _ := reader.ReadString('\n')
		str = strings.TrimSpace(str)

		value, err := strconv.ParseFloat(str, 64)
		if err == nil {
			return value
		}

		fmt.Println("Invalid number, please try again.")
	}
}

//
// ─── READ PASSWORD WITH STAR MASKING ───────────────────────────────────────────
//

// ReadPassword reads a password while showing '*' for each character
func ReadPasswordMasked(prompt string) (string, error) {
	fmt.Print(prompt)

	fd := int(os.Stdin.Fd())

	// Save original terminal state globally for signal handler use
	var err error
	originalState, err = term.GetState(fd)
	if err != nil {
		return "", err
	}

	// Enter raw mode
	rawState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}

	// Always restore terminal at end
	defer term.Restore(fd, rawState)

	buf := make([]byte, 0, 32)
	var b = make([]byte, 1)

	for {
		_, err := os.Stdin.Read(b)
		if err != nil {
			return "", err
		}

		ch := b[0]

		// ENTER or RETURN → finish
		if ch == 13 || ch == 10 {
			fmt.Println()
			break
		}

		// BACKSPACE or DEL
		if ch == 127 || ch == 8 {
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Print("\b \b") // erase last '*'
			}
			continue
		}

		// Ignore control characters
		if ch < 32 || ch > 126 {
			continue
		}

		// Normal character
		buf = append(buf, ch)
		fmt.Print("*")
	}

	return string(buf), nil
}

// ReadChoice asks the user to choose a menu option within a valid range
func ReadChoice(prompt string, min, max int) int {
	for {
		fmt.Print(prompt)
		str, _ := reader.ReadString('\n')
		str = strings.TrimSpace(str)

		value, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please input a valid choice!")
			continue
		}

		if value < min || value > max {
			fmt.Println("Please input a valid choice!")
			continue
		}

		return value
	}
}

func ReadConfirmation(prompt string) bool {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" {
			return true
		}
		if input == "n" {
			return false
		}

		fmt.Println("Invalid input! Please enter 'y' or 'n'.")
	}
}

func ReadDate(prompt string) string {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		_, err := time.Parse("2006-01-02", input)
		if err != nil {
			fmt.Println("Invalid date format! Use YYYY-MM-DD.")
			continue
		}

		return input
	}
}

type PageInput struct {
	Command string // "<" or ">"
	ID      int    // product ID (if any)
}

// ReadPagingInput handles "<", ">", or a product ID (0 = Back)
func ReadPagingInput(prompt string) PageInput {
	for {
		fmt.Print(prompt)
		str, _ := reader.ReadString('\n')
		str = strings.TrimSpace(str)

		// Commands
		if str == "<" || str == ">" {
			return PageInput{Command: str}
		}

		// numeric input (allow 0 as "Back")
		id, err := strconv.Atoi(str)
		if err == nil && id >= 0 {
			return PageInput{ID: id}
		}

		fmt.Println("Please input '<', '>', 0 (Back) or a valid product ID.")
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ReadEmail(prompt string) string {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(prompt)

		if !scanner.Scan() {
			continue
		}

		email := strings.TrimSpace(scanner.Text())

		if emailRegex.MatchString(email) {
			return email
		}

		fmt.Println("Invalid email format. Please try again.")
	}
}

func ParseIntList(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" {
		return []int{}
	}

	parts := strings.Split(s, ",")
	var result []int

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if n, err := strconv.Atoi(p); err == nil {
			result = append(result, n)
		}
	}
	return result
}
