package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ANSI color codes
const (
	colorRed        = "\033[31m"
	colorGreen      = "\033[32m"
	colorYellow     = "\033[33m"
	colorBlue       = "\033[34m"
	colorMagenta    = "\033[35m"
	colorCyan       = "\033[36m"
	colorReset      = "\033[0m"
	// Gradient colors for hangman
	colorLightYellow = "\033[38;5;226m"  // Very light yellow
	colorYellowOrange = "\033[38;5;220m" // Yellow-orange
	colorOrange      = "\033[38;5;214m"  // Orange
	colorDarkOrange  = "\033[38;5;208m"  // Dark orange
	colorLightRed    = "\033[38;5;203m"  // Light red
	colorDarkRed     = "\033[38;5;196m"  // Dark red
)

// Color helper functions
func red(text string) string {
	return colorRed + text + colorReset
}

func green(text string) string {
	return colorGreen + text + colorReset
}

func yellow(text string) string {
	return colorYellow + text + colorReset
}

func blue(text string) string {
	return colorBlue + text + colorReset
}

func magenta(text string) string {
	return colorMagenta + text + colorReset
}

func cyan(text string) string {
	return colorCyan + text + colorReset
}

// Gradient color helper function
func colorByAttempts(text string, attemptsLeft int) string {
	switch attemptsLeft {
	case 6:
		return colorLightYellow + text + colorReset
	case 5:
		return colorYellowOrange + text + colorReset
	case 4:
		return colorOrange + text + colorReset
	case 3:
		return colorDarkOrange + text + colorReset
	case 2:
		return colorLightRed + text + colorReset
	case 1:
		return colorRed + text + colorReset
	case 0:
		return colorDarkRed + text + colorReset
	default:
		return text
	}
}

// Structs for game state and player scores
type GameState struct {
	WordToGuess      string
	GuessedLetters   []string
	AttemptsLeft     int
	Player1          string
	Player2          string
	IsMultiplayer    bool
}

type Scoreboard struct {
	PlayerStats map[string]int
	TotalGames  map[string]int
}

func mainMenu() string {
	fmt.Println(cyan("\n=== Hangman Game ==="))
	fmt.Println(yellow("1. Start New Game"))
	fmt.Println(yellow("2. Continue Saved Game"))
	fmt.Println(yellow("3. View Scoreboard"))
	fmt.Println(red("4. Exit"))
	fmt.Print(green("Enter your choice (1-4): "))
	var choice string
	fmt.Scanln(&choice)
	return choice
}

func startNewGame() GameState {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Single Player Mode or Multiplayer Mode? (Enter 'S' or 'M')")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(strings.ToUpper(mode))
	var player1 string
	var word string
	var player2 string = ""
	isMultiplayer := mode == "M"

	if isMultiplayer {
		fmt.Println("\n=== New Game ===")
		fmt.Println("Enter Player 1 name:")
		player1, _ = reader.ReadString('\n')
		player1 = strings.TrimSpace(player1)
		fmt.Println("Enter Player 2 name:")
		player2, _ = reader.ReadString('\n')
		player2 = strings.TrimSpace(player2)
		fmt.Println(player1 + ", enter a word for " + player2 + " to guess:")
		word, _ = reader.ReadString('\n')
		word = strings.TrimSpace(word)
		clearScreen()
		fmt.Println("Word entered. Handing over to " + player2 + ".")
	} else {
		fmt.Println("\n=== New Game ===")
		fmt.Println("Enter Player 1 name:")
		player1, _ = reader.ReadString('\n')
		player1 = strings.TrimSpace(player1)
		word = generateRandomWord()
	}

	return GameState{
		WordToGuess:    strings.ToUpper(word),
		GuessedLetters: []string{},
		AttemptsLeft:   6,
		Player1:        player1,
		Player2:        player2,
		IsMultiplayer:  isMultiplayer,
	}
}

func generateRandomWord() string {
	words := []string{"COMPUTER", "PROGRAMMER", "HANGMAN", "SOFTWARE", "DEVELOPER"}
	rand.Seed(time.Now().UnixNano())
	return words[rand.Intn(len(words))]
}

func playGame(game *GameState) {
	reader := bufio.NewReader(os.Stdin)
	for game.AttemptsLeft > 0 {
		displayGameStatus(game)
		fmt.Println("\nEnter your guess (or type 'save' to save the game):")
		guess, _ := reader.ReadString('\n')
		guess = strings.TrimSpace(strings.ToUpper(guess))

		if guess == "SAVE" {
			saveGame(game)
			fmt.Println("Game saved. Exiting to main menu.")
			return
		}

		if len(guess) != 1 || strings.ContainsAny(guess, "0123456789") {
			fmt.Println("Invalid input! Please enter a single letter.")
			continue
		}

		game.GuessedLetters = append(game.GuessedLetters, guess)

		if strings.Contains(game.WordToGuess, guess) {
			fmt.Println("Correct guess!")
		} else {
			game.AttemptsLeft--
			fmt.Println("Incorrect guess!")
		}

		if isWordGuessed(game) {
			displayGameStatus(game)
			fmt.Println("\nCongratulations! You've guessed the word:", game.WordToGuess)
			if game.IsMultiplayer && game.Player2 != "" {
				updateScoreboard(game.Player2, true)  // Update scoreboard when player 2 wins
			} else if !game.IsMultiplayer && game.Player1 != "" {
				updateScoreboard(game.Player1, true)  // Update scoreboard when player 1 wins
			}
			return
		}
	}

	displayGameStatus(game)  // Show final state
	fmt.Println("\nGame Over! The word was:", game.WordToGuess)
	if game.IsMultiplayer && game.Player2 != "" {
		updateScoreboard(game.Player2, false)  // Update total games for player 2 on loss
	} else if !game.IsMultiplayer && game.Player1 != "" {
		updateScoreboard(game.Player1, false)  // Update total games for player 1 on loss
	}
}

func isWordGuessed(game *GameState) bool {
	for _, char := range game.WordToGuess {
		if !contains(game.GuessedLetters, string(char)) {
			return false
		}
	}
	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func saveGame(game *GameState) {
	file, _ := os.Create("saved_game.json")
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.Encode(game)
	fmt.Println("Game saved successfully!")
}

func loadGame() GameState {
	file, _ := os.Open("saved_game.json")
	defer file.Close()
	var game GameState
	decoder := json.NewDecoder(file)
	decoder.Decode(&game)
	fmt.Println("Game loaded successfully!")
	return game
}

func viewScoreboard() {
	file, _ := os.Open("scoreboard.json")
	defer file.Close()
	var scoreboard Scoreboard
	decoder := json.NewDecoder(file)
	decoder.Decode(&scoreboard)

	fmt.Println("\n=== Scoreboard ===")
	// Iterate over TotalGames instead of PlayerStats to show all players
	for player, totalGames := range scoreboard.TotalGames {
		wins := scoreboard.PlayerStats[player] // This will be 0 if player has no wins
		fmt.Printf("%s: %d wins, %d total games\n", player, wins, totalGames)
	}
}

func updateScoreboard(player string, isWin bool) {
	if player == "" {
		return  // Don't update scoreboard if no player name is provided
	}
	scoreboard := Scoreboard{
		PlayerStats: make(map[string]int),
		TotalGames: make(map[string]int),
	}
	file, err := os.OpenFile("scoreboard.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening scoreboard:", err)
		return
	}
	defer file.Close()

	// Try to read existing scoreboard
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&scoreboard); err != nil {
		// If file is empty or invalid, we already have an initialized scoreboard
		// with empty maps from above
	}

	// Update score
	if isWin {
		scoreboard.PlayerStats[player]++
	}
	scoreboard.TotalGames[player]++

	// Clear file and write updated scoreboard
	file.Seek(0, 0)
	file.Truncate(0)
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(scoreboard); err != nil {
		fmt.Println("Error updating scoreboard:", err)
		return
	}
}

func displayGameStatus(game *GameState) {
	clearScreen()
	fmt.Println(cyan("\n=== Hangman Game ==="))
	fmt.Printf("%s %s\n", magenta("Player:"), blue(game.Player1))
	if game.IsMultiplayer {
		fmt.Printf("%s %s\n", magenta("Word was provided by:"), blue(game.Player2))
	}
	fmt.Printf("\n%s %s\n", magenta("Attempts Left:"), red(fmt.Sprintf("%d", game.AttemptsLeft)))
	fmt.Printf("%s %s\n", magenta("Word to Guess:"), blue(getWordDisplay(game.WordToGuess, game.GuessedLetters)))
	fmt.Printf("%s %s\n", magenta("Guessed Letters:"), green(strings.Join(game.GuessedLetters, " ")))
	displayHangmanArt(game.AttemptsLeft)
}

// getWordDisplay generates the current word display with underscores for unguessed letters.
func getWordDisplay(word string, guessedLetters []string) string {
	display := ""
	for _, char := range word {
		if contains(guessedLetters, string(char)) {
			display += string(char) + " "
		} else {
			display += "_ "
		}
	}
	return display
}

func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default: // For macOS and Linux
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func displayHangmanArt(attemptsLeft int) {
	hangmanStages := []string{
		`
   +---+
   |   |
       |
       |
       |
       |
=========`,
		`
   +---+
   |   |
   O   |
       |
       |
       |
=========`,
		`
   +---+
   |   |
   O   |
   |   |
       |
       |
=========`,
		`
   +---+
   |   |
   O   |
  /|   |
       |
       |
=========`,
		`
   +---+
   |   |
   O   |
  /|\  |
       |
       |
=========`,
		`
   +---+
   |   |
   O   |
  /|\  |
  /    |
       |
=========`,
		`
   +---+
   |   |
   O   |
  /|\  |
  / \  |
       |
=========`,
	}
	
	// Select the appropriate Hangman stage based on attempts left
	if attemptsLeft < 0 {
		attemptsLeft = 0
	}
	fmt.Println(colorByAttempts(hangmanStages[len(hangmanStages)-1-attemptsLeft], attemptsLeft))
}

func main() {
	for {
		choice := mainMenu()
		switch choice {
		case "1":
			game := startNewGame()
			playGame(&game)
		case "2":
			game := loadGame()
			playGame(&game)
		case "3":
			viewScoreboard()
		case "4":
			fmt.Println(green("\nThank you for playing! Goodbye!"))
			return
		default:
			fmt.Println(red("Invalid choice. Please try again."))
		}
	}
}

func promptSaveGame(game *GameState) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(yellow("\nDo you want to save the game? (Y/N)"))
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToUpper(choice))

	if choice == "Y" {
		saveGame(game)
	}
}

func promptLoadGame() *GameState {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(yellow("\nDo you want to load the saved game? (Y/N)"))
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToUpper(choice))

	if choice == "Y" {
		game := loadGame()
		return &game
	}
	return nil
}
