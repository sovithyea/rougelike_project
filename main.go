package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

// Screen / map dimensions
const (
	ScreenWidth  = 80
	ScreenHeight = 50
	MapWidth     = 80
	MapHeight    = 45
)

// Tile colours
var (
	colorDarkWall   = tcell.NewRGBColor(0, 0, 100)
	colorDarkGround = tcell.NewRGBColor(50, 50, 150)
)

// Tile

// Tile is a single cell of the map.
type Tile struct {
	Blocked    bool
	BlockSight bool
}

func emptyTile() Tile { return Tile{Blocked: false, BlockSight: false} }
func wallTile() Tile  { return Tile{Blocked: true, BlockSight: true} }

// Map

// Map is a 2-D grid of Tiles, indexed [x][y].
type Map [][]Tile

func makeMap() Map {
	// Fill every cell with an empty (walkable) tile
	m := make(Map, MapWidth)
	for x := range m {
		m[x] = make([]Tile, MapHeight)
		for y := range m[x] {
			m[x][y] = emptyTile()
		}
	}

	// Two test pillars so we can see collision working
	m[30][22] = wallTile()
	m[50][22] = wallTile()

	return m
}

// Game (holds all persistent game state)

type Game struct {
	Map Map
}

// Object  (player, NPC, monster, item — anything on the map)

type Object struct {
	X, Y  int
	Char  rune
	Color tcell.Color
}

func NewObject(x, y int, ch rune, color tcell.Color) *Object {
	return &Object{X: x, Y: y, Char: ch, Color: color}
}

// MoveBy moves the object by (dx, dy) if the destination tile is not blocked.
func (o *Object) MoveBy(dx, dy int, game *Game) {
	nx, ny := o.X+dx, o.Y+dy
	if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
		if !game.Map[nx][ny].Blocked {
			o.X, o.Y = nx, ny
		}
	}
}

// Draw renders the object onto the screen.
func (o *Object) Draw(screen tcell.Screen) {
	style := tcell.StyleDefault.
		Foreground(o.Color).
		Background(tcell.ColorBlack)
	screen.SetContent(o.X, o.Y, o.Char, nil, style)
}

// Input handling

// handleKeys processes a key event. Returns true if the game should quit.
func handleKeys(ev *tcell.EventKey, game *Game, player *Object) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		return true
	case tcell.KeyUp:
		player.MoveBy(0, -1, game)
	case tcell.KeyDown:
		player.MoveBy(0, 1, game)
	case tcell.KeyLeft:
		player.MoveBy(-1, 0, game)
	case tcell.KeyRight:
		player.MoveBy(1, 0, game)
	}
	return false
}

// Rendering

// renderAll draws the map and then every object on top.
func renderAll(screen tcell.Screen, game *Game, objects []*Object) {
	// Draw map tiles
	for x := 0; x < MapWidth; x++ {
		for y := 0; y < MapHeight; y++ {
			wall := game.Map[x][y].BlockSight
			var bg tcell.Color
			if wall {
				bg = colorDarkWall
			} else {
				bg = colorDarkGround
			}
			style := tcell.StyleDefault.Background(bg)
			screen.SetContent(x, y, ' ', nil, style)
		}
	}

	// Draw all objects on top of the map
	for _, obj := range objects {
		obj.Draw(screen)
	}
}

// Main

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to create screen: %v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to init screen: %v", err)
	}
	defer screen.Fini()

	game := &Game{Map: makeMap()}

	// Create player and a yellow NPC to test the object system
	player := NewObject(MapWidth/2, MapHeight/2, '@', tcell.ColorWhite)
	npc := NewObject(MapWidth/2-5, MapHeight/2, '@', tcell.ColorYellow)
	objects := []*Object{player, npc}

	// Main game loop
	for {
		screen.Clear()
		renderAll(screen, game, objects)
		screen.Show()

		ev := screen.PollEvent()
		switch e := ev.(type) {
		case *tcell.EventKey:
			if handleKeys(e, game, player) {
				return
			}
		case *tcell.EventResize:
			screen.Sync()
		}
	}
}