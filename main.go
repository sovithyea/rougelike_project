package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

// Screen dimensions (in characters)
const (
	ScreenWidth  = 80
	ScreenHeight = 50
)

func main() {
	// Initialize the screen
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to create screen: %v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to init screen: %v", err)
	}
	defer screen.Fini()

	// Style for drawing the player
	playerStyle := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack)

	// Player starts in the centre of the screen
	playerX := ScreenWidth / 2
	playerY := ScreenHeight / 2

	// Main game loop
	for {
		// Clear the screen
		screen.Clear()

		// Draw the player '@' at current position
		screen.SetContent(playerX, playerY, '@', nil, playerStyle)

		// Flush everything to the terminal
		screen.Show()

		// Wait for input
		ev := screen.PollEvent()
		switch e := ev.(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyEscape:
				// Exit the game
				return
			case tcell.KeyUp:
				playerY--
			case tcell.KeyDown:
				playerY++
			case tcell.KeyLeft:
				playerX--
			case tcell.KeyRight:
				playerX++
			}

			// Keep player within screen bounds
			if playerX < 0 {
				playerX = 0
			}
			if playerX >= ScreenWidth {
				playerX = ScreenWidth - 1
			}
			if playerY < 0 {
				playerY = 0
			}
			if playerY >= ScreenHeight {
				playerY = ScreenHeight - 1
			}

		case *tcell.EventResize:
			// Handle terminal resize
			screen.Sync()
		}
	}
}