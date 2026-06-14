package main

import (
	"log"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

// Screen / map dimensions
const (
	ScreenWidth  = 80
	ScreenHeight = 50
	MapWidth     = 80
	MapHeight    = 45
)

// Dungeon generator parameters
const (
	RoomMaxSize = 10
	RoomMinSize = 6
	MaxRooms    = 30
)

// Tile colours
var (
	colorDarkWall   = tcell.NewRGBColor(0, 0, 100)
	colorDarkGround = tcell.NewRGBColor(50, 50, 150)
)

// Tile

type Tile struct {
	Blocked    bool
	BlockSight bool
}

func emptyTile() Tile { return Tile{false, false} }
func wallTile() Tile  { return Tile{true, true} }

// Rect  (a rectangular room)

type Rect struct {
	X1, Y1, X2, Y2 int
}

func NewRect(x, y, w, h int) Rect {
	return Rect{x, y, x + w, y + h}
}

func (r Rect) Center() (int, int) {
	return (r.X1 + r.X2) / 2, (r.Y1 + r.Y2) / 2
}

func (r Rect) Intersects(other Rect) bool {
	return r.X1 <= other.X2 && r.X2 >= other.X1 &&
		r.Y1 <= other.Y2 && r.Y2 >= other.Y1
}

// Map

type Map [][]Tile

// createRoom carves out a room by setting its interior tiles to empty.
func createRoom(room Rect, m Map) {
	for x := room.X1 + 1; x < room.X2; x++ {
		for y := room.Y1 + 1; y < room.Y2; y++ {
			m[x][y] = emptyTile()
		}
	}
}

func createHTunnel(x1, x2, y int, m Map) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		m[x][y] = emptyTile()
	}
}

func createVTunnel(y1, y2, x int, m Map) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		m[x][y] = emptyTile()
	}
}

// makeMap builds a procedurally generated dungeon.
// It positions the player at the centre of the first room.
func makeMap(player *Object) Map {
	// Start with a solid wall everywhere
	m := make(Map, MapWidth)
	for x := range m {
		m[x] = make([]Tile, MapHeight)
		for y := range m[x] {
			m[x][y] = wallTile()
		}
	}

	var rooms []Rect

	for i := 0; i < MaxRooms; i++ {
		w := RoomMinSize + rand.Intn(RoomMaxSize-RoomMinSize+1)
		h := RoomMinSize + rand.Intn(RoomMaxSize-RoomMinSize+1)
		x := rand.Intn(MapWidth - w - 1)
		y := rand.Intn(MapHeight - h - 1)

		newRoom := NewRect(x, y, w, h)

		// Reject the room if it overlaps any existing room
		overlaps := false
		for _, other := range rooms {
			if newRoom.Intersects(other) {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}

		// Carve the room into the map
		createRoom(newRoom, m)
		cx, cy := newRoom.Center()

		if len(rooms) == 0 {
			// First room — place the player here
			player.X, player.Y = cx, cy
		} else {
			// Connect this room to the previous one with tunnels
			prevCX, prevCY := rooms[len(rooms)-1].Center()
			if rand.Intn(2) == 0 {
				createHTunnel(prevCX, cx, prevCY, m)
				createVTunnel(prevCY, cy, cx, m)
			} else {
				createVTunnel(prevCY, cy, prevCX, m)
				createHTunnel(prevCX, cx, cy, m)
			}
		}

		rooms = append(rooms, newRoom)
	}

	return m
}

// Game

type Game struct {
	Map Map
}

// Object

type Object struct {
	X, Y  int
	Char  rune
	Color tcell.Color
}

func NewObject(x, y int, ch rune, color tcell.Color) *Object {
	return &Object{X: x, Y: y, Char: ch, Color: color}
}

func (o *Object) MoveBy(dx, dy int, game *Game) {
	nx, ny := o.X+dx, o.Y+dy
	if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
		if !game.Map[nx][ny].Blocked {
			o.X, o.Y = nx, ny
		}
	}
}

func (o *Object) Draw(screen tcell.Screen) {
	style := tcell.StyleDefault.
		Foreground(o.Color).
		Background(tcell.ColorBlack)
	screen.SetContent(o.X, o.Y, o.Char, nil, style)
}

// Input

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

func renderAll(screen tcell.Screen, game *Game, objects []*Object) {
	for x := 0; x < MapWidth; x++ {
		for y := 0; y < MapHeight; y++ {
			var bg tcell.Color
			if game.Map[x][y].BlockSight {
				bg = colorDarkWall
			} else {
				bg = colorDarkGround
			}
			screen.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(bg))
		}
	}
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

	// Player starts at (0,0) — makeMap will move them to the first room
	player := NewObject(0, 0, '@', tcell.ColorWhite)
	objects := []*Object{player}

	game := &Game{Map: makeMap(player)}

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