package entities

type Enemy struct {
	*Sprite
	Health  uint16
	Attack  uint16
	Defense uint16
	Speed   uint16
}
