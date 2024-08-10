package entities

import "github.com/hajimehoshi/ebiten/v2"

// base type for all entities/sprites
type Sprite struct {
	X, Y  float64 //position
	Image *ebiten.Image
}
