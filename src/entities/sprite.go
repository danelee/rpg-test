package entities

import "github.com/hajimehoshi/ebiten/v2"

// base type for all entities/sprites
type Sprite struct {
	X, Y  float64 //position
	Image *ebiten.Image
}

type Projectile struct {
	*Sprite
	Speed   uint16
	Damage  uint16
	Impact  bool
	Enabled bool
}
