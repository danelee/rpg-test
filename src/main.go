package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand/v2"

	"github.com/danelee/rpg-test/src/entities"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// const
const SCREEN_WIDTH int = 640
const SCREEN_HEIGHT int = 480

// Game implements ebiten.Game interface.
type Game struct {
	player       *entities.Player
	enemies      []*entities.Enemy
	camera       *Camera
	tilemap      *Tilemap
	tilemapImage *ebiten.Image
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write game's logic update.
	g.movePlayer()

	for _, enemy := range g.enemies {
		g.followPlayer(enemy)
	}

	//if ebiten.CursorMode() == 0 {
	//	g.followCursor()
	//}

	g.camera.Follow(g.player.X+8, g.player.Y+8, float64(SCREEN_WIDTH), float64(SCREEN_HEIGHT))
	g.camera.Constrain(
		float64(g.tilemap.Layers[0].Width)*16.0,
		float64(g.tilemap.Layers[0].Height)*16.0,
		float64(SCREEN_WIDTH),
		float64(SCREEN_HEIGHT))

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {

	// Write your game's rendering.
	screen.Fill(color.RGBA{192, 192, 192, 127})

	opts := ebiten.DrawImageOptions{}

	//TileMap rendering

	for _, layer := range g.tilemap.Layers {
		for index, id := range layer.Data {
			x := index % layer.Width
			y := index / layer.Height

			x *= 16
			y *= 16

			srcX := (id - 1) % 22
			srcY := (id - 1) / 22

			srcX *= 16
			srcY *= 16

			opts.GeoM.Translate(float64(x), float64(y))

			opts.GeoM.Translate(g.camera.X, g.camera.Y)

			screen.DrawImage(
				g.tilemapImage.SubImage(image.Rect(srcX, srcY, srcX+16, srcY+16)).(*ebiten.Image),
				&opts,
			)
			opts.GeoM.Reset()
		}
	}

	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.camera.X, g.camera.Y)

	//draw player
	drawPlayer(g.player, screen, &opts)
	//screen.DrawImage(g.player.Image.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), &opts)

	//draw enemies
	for _, enemy := range g.enemies {
		opts.GeoM.Reset()
		opts.GeoM.Translate(enemy.X, enemy.Y)

		opts.GeoM.Translate(g.camera.X, g.camera.Y)
		screen.DrawImage(enemy.Image.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), &opts)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("enemy X = %v, enemy Y = %v", enemy.X, enemy.Y), 5, 5)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%v", SCREEN_WIDTH-g.tilemap.Layers[0].Width*16.0), 50, 50)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ebiten.WindowSize()
}

func main() {
	// Specify the window size as you like. Here,  a doubled size is specified.
	ebiten.SetWindowSize(SCREEN_WIDTH, SCREEN_HEIGHT)
	ebiten.SetWindowTitle("Man vs Wild")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	player, err := loadPlayer()
	if err != nil {
		log.Fatal(err)
	}

	enemy, err := loadEnemy()
	if err != nil {
		log.Fatal(err)
	}

	tilemap, err := NewTilemap("assets/maps/testmap.json")
	if err != nil {
		log.Fatal(err)
	}

	tilemapImage, _, err := ebitenutil.NewImageFromFile("assets/images/TilesetFloor.png")
	if err != nil {
		log.Fatal(err)
	}

	// Game struct everything associated with the game
	game := &Game{
		player: player,
		enemies: []*entities.Enemy{
			enemy,
		},
		camera:       NewCamera(0, 0),
		tilemap:      tilemap,
		tilemapImage: tilemapImage,
	}

	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

//player functionality

// load player
func loadPlayer() (*entities.Player, error) {
	player := &entities.Player{
		Name:    "Hero",
		Health:  13,
		Attack:  10,
		Defense: 7,
		Speed:   3,
		Sprite: &entities.Sprite{
			X: float64(SCREEN_WIDTH / 2),
			Y: float64(SCREEN_HEIGHT / 2),
		},
	}
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/Hunter/SpriteSheet.png")
	if err != nil {
		return nil, err
	}

	player.Image = playerImg

	return player, nil
}

// add player image based on direction
func drawPlayer(player *entities.Player, screen *ebiten.Image, opts *ebiten.DrawImageOptions) {
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		screen.DrawImage(player.Image.SubImage(image.Rect(48, 0, 64, 16)).(*ebiten.Image), opts)

	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		screen.DrawImage(player.Image.SubImage(image.Rect(32, 0, 48, 16)).(*ebiten.Image), opts)
	} else if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		screen.DrawImage(player.Image.SubImage(image.Rect(16, 0, 32, 16)).(*ebiten.Image), opts)
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		screen.DrawImage(player.Image.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), opts)
	} else {
		screen.DrawImage(player.Image.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), opts)
	}
}

// move player
func (g *Game) movePlayer() {
	tilemapWidthPx := g.tilemap.Layers[0].Width * 16
	tilemapHeightPx := g.tilemap.Layers[0].Height * 16
	player := g.player
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		if player.X+float64(player.Speed) > float64(tilemapWidthPx-16) {
			player.X = float64(tilemapWidthPx - 16)
		} else {
			player.X += float64(player.Speed)
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		if player.X-float64(player.Speed) < 0 {
			player.X = 0
		} else {
			player.X -= float64(player.Speed)
		}

	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		if player.Y-float64(player.Speed) < 0 {
			player.Y = 0
		} else {
			player.Y -= float64(player.Speed)
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		if player.Y+float64(player.Speed) > float64(tilemapHeightPx-16) {
			player.Y = float64(tilemapHeightPx - 16)
		} else {
			player.Y += float64(player.Speed)
		}
	}

}

func (g *Game) followCursor() {
	cursorX, cursorY := ebiten.CursorPosition()
	if cursorX > SCREEN_WIDTH || cursorX < 0 || cursorY > SCREEN_HEIGHT || cursorY < 0 {
		return
	}

	if g.player.X < float64(cursorX) {
		g.player.X += float64(g.player.Speed)
	}
	if g.player.X > float64(cursorX) {
		g.player.X -= float64(g.player.Speed)
	}
	if g.player.Y < float64(cursorY) {
		g.player.Y += float64(g.player.Speed)
	}
	if g.player.Y > float64(cursorY) {
		g.player.Y -= float64(g.player.Speed)
	}

}

// load sprites
func loadEnemy() (*entities.Enemy, error) {
	image, _, err := ebitenutil.NewImageFromFile("assets/images/Cyclope/SpriteSheet.png")
	if err != nil {
		return nil, err
	}
	enemy := &entities.Enemy{
		Health:  20,
		Attack:  3,
		Defense: 2,
		Speed:   2,
		Sprite: &entities.Sprite{
			X:     (rand.Float64() * float64(SCREEN_WIDTH)),
			Y:     (rand.Float64() * float64(SCREEN_HEIGHT)),
			Image: image,
		},
	}

	return enemy, nil
}

func (g *Game) followPlayer(enemy *entities.Enemy) {
	if enemy.X < g.player.X {
		enemy.X += float64(enemy.Speed)
	}
	if enemy.Y < g.player.Y {
		enemy.Y += float64(enemy.Speed)
	}
	if enemy.X > g.player.X {
		enemy.X -= float64(enemy.Speed)
	}
	if enemy.Y > g.player.Y {
		enemy.Y -= float64(enemy.Speed)
	}
}
