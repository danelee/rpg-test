package entities

type Player struct {
	*Sprite
	Name    string
	Health  uint16
	Attack  uint16
	Defense uint16
	Speed   uint16
}
