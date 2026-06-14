# Roguelike

A classic ASCII dungeon-crawler roguelike written in Go, built by following the [Complete Roguelike Tutorial](https://tomassedovic.github.io/roguelike-tutorial/) (originally Rust + libtcod, ported here to Go + tcell).

## Features (in progress)

- [x] Part 1 — Graphics & movement
- [x] Part 2 — Object system & map
- [x] Part 3 — Dungeon generator (In orogress)
- [ ] Part 4 — Field of view & fog of war
- [ ] Part 5 — Preparing for combat
- [ ] Part 6 — Combat
- [ ] Part 7 — GUI
- [ ] Part 8 — Items & inventory
- [ ] Part 9 — Spells & ranged combat
- [ ] Part 10 — Main menu & saving
- [ ] Part 11 — Dungeon levels & character progression
- [ ] Part 12 — Monster & item progression
- [ ] Part 13 — Adventure gear

## Requirements

- Go 1.21+

## Running

```bash
go mod tidy
go run .
```

## Controls

| Key         | Action       |
|-------------|--------------|
| Arrow keys  | Move player  |
| Escape      | Quit         |

## Dependencies

- [tcell](https://github.com/gdamore/tcell) — terminal rendering & input