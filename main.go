package main

import (
	"log"
	"math"
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

// FOV parameters
const TorchRadius = 10

// Colours: dark (outside FOV) and light (inside FOV)
var (
	colorDarkWall    = tcell.NewRGBColor(0, 0, 100)
	colorLightWall   = tcell.NewRGBColor(130, 110, 50)
	colorDarkGround  = tcell.NewRGBColor(50, 50, 150)
	colorLightGround = tcell.NewRGBColor(200, 180, 50)
)

// -----------------------------------------------------------------------
// Tile
// -----------------------------------------------------------------------

type Tile struct {
	Blocked    bool
	BlockSight bool
	Explored   bool
}

func emptyTile() Tile { return Tile{Blocked: false, BlockSight: false, Explored: false} }
func wallTile() Tile  { return Tile{Blocked: true, BlockSight: true, Explored: false} }

// -----------------------------------------------------------------------
// FOV  (ray-casting)
// -----------------------------------------------------------------------

// computeFOV returns a 2-D boolean grid of which tiles are visible from
// (originX, originY) within radius tiles, respecting BlockSight walls.
func computeFOV(m Map, originX, originY, radius int) [][]bool {
	visible := make([][]bool, MapWidth)
	for x := range visible {
		visible[x] = make([]bool, MapHeight)
	}

	// Cast rays at many angles
	steps := 360 * 4 // fine enough for a grid
	for i := 0; i < steps; i++ {
		angle := float64(i) * math.Pi * 2 / float64(steps)
		dx := math.Cos(angle)
		dy := math.Sin(angle)

		rx, ry := float64(originX)+0.5, float64(originY)+0.5
		for dist := 0; dist < radius; dist++ {
			x, y := int(rx), int(ry)
			if x < 0 || x >= MapWidth || y < 0 || y >= MapHeight {
				break
			}
			visible[x][y] = true
			if m[x][y].BlockSight {
				break // wall — stop this ray but mark the wall visible
			}
			rx += dx
			ry += dy
		}
	}
	return visible
}

// -----------------------------------------------------------------------
// Rect
// -----------------------------------------------------------------------

type Rect struct{ X1, Y1, X2, Y2 int }

func NewRect(x, y, w, h int) Rect { return Rect{x, y, x + w, y + h} }

func (r Rect) Center() (int, int) { return (r.X1 + r.X2) / 2, (r.Y1 + r.Y2) / 2 }

func (r Rect) Intersects(o Rect) bool {
	return r.X1 <= o.X2 && r.X2 >= o.X1 && r.Y1 <= o.Y2 && r.Y2 >= o.Y1
}

// -----------------------------------------------------------------------
// Map
// -----------------------------------------------------------------------

type Map [][]Tile

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

func makeMap(player *Object) Map {
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

		createRoom(newRoom, m)
		cx, cy := newRoom.Center()

		if len(rooms) == 0 {
			player.X, player.Y = cx, cy
		} else {
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

// -----------------------------------------------------------------------
// Game
// -----------------------------------------------------------------------

type Game struct {
	Map Map
}

// -----------------------------------------------------------------------
// Object
// -----------------------------------------------------------------------

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

// -----------------------------------------------------------------------
// Input
// -----------------------------------------------------------------------

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

// -----------------------------------------------------------------------
// Rendering
// -----------------------------------------------------------------------

func renderAll(screen tcell.Screen, game *Game, objects []*Object, visible [][]bool) {
	for x := 0; x < MapWidth; x++ {
		for y := 0; y < MapHeight; y++ {
			isVisible := visible[x][y]
			isWall := game.Map[x][y].BlockSight

			if isVisible {
				game.Map[x][y].Explored = true
			}

			if !game.Map[x][y].Explored {
				continue // still in the dark — don't draw at all
			}

			var bg tcell.Color
			switch {
			case isVisible && isWall:
				bg = colorLightWall
			case isVisible && !isWall:
				bg = colorLightGround
			case !isVisible && isWall:
				bg = colorDarkWall
			default:
				bg = colorDarkGround
			}
			screen.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(bg))
		}
	}

	// Only draw objects the player can currently see
	for _, obj := range objects {
		if visible[obj.X][obj.Y] {
			obj.Draw(screen)
		}
	}
}

// -----------------------------------------------------------------------
// Main
// -----------------------------------------------------------------------

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to create screen: %v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to init screen: %v", err)
	}
	defer screen.Fini()

	player := NewObject(0, 0, '@', tcell.ColorWhite)
	objects := []*Object{player}
	game := &Game{Map: makeMap(player)}

	// Compute initial FOV
	visible := computeFOV(game.Map, player.X, player.Y, TorchRadius)
	prevX, prevY := player.X, player.Y

	for {
		screen.Clear()
		renderAll(screen, game, objects, visible)
		screen.Show()

		ev := screen.PollEvent()
		switch e := ev.(type) {
		case *tcell.EventKey:
			if handleKeys(e, game, player) {
				return
			}
			// Recompute FOV only when the player actually moved
			if player.X != prevX || player.Y != prevY {
				visible = computeFOV(game.Map, player.X, player.Y, TorchRadius)
				prevX, prevY = player.X, player.Y
			}
		case *tcell.EventResize:
			screen.Sync()
		}
	}
}
