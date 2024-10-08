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

// Game implements ebiten.Game interface.
type Game struct {
	player  *entities.Player
	enemies []*entities.Enemy
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write game's logic update.
	movePlayer(g.player)

	for _, enemy := range g.enemies {
		g.followPlayer(enemy)
	}

	if ebiten.CursorMode() == 0 {
		g.followCursor()
	}

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {

	// Write your game's rendering.
	screen.Fill(color.RGBA{192, 192, 192, 127})

	opts := ebiten.DrawImageOptions{}
	opts.GeoM.Translate(g.player.X, g.player.Y)

	//draw player
	screen.DrawImage(g.player.Image.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), &opts)

	//draw enemies
	for _, enemy := range g.enemies {
		opts.GeoM.Reset()
		opts.GeoM.Translate(enemy.X, enemy.Y)

		screen.DrawImage(enemy.Image.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image), &opts)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("enemy X = %v, enemy Y = %v", enemy.X, enemy.Y), 5, 5)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ebiten.WindowSize()
}

func main() {
	// Specify the window size as you like. Here,  a doubled size is specified.
	ebiten.SetWindowSize(640, 480)
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

	// Game struct everything associated with the game
	game := &Game{
		player: player,
		enemies: []*entities.Enemy{
			enemy,
		},
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
		Speed:   5,
		Sprite: &entities.Sprite{
			X: 10,
			Y: 10,
		},
	}
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/Hunter/SpriteSheet.png")
	if err != nil {
		return nil, err
	}

	player.Image = playerImg

	return player, nil
}

// move player
func movePlayer(player *entities.Player) {
	if (ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && (ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)) {
		player.Y -= float64(player.Speed) * 0.75
		player.X -= float64(player.Speed) * 0.75
	} else if (ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && (ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)) {
		player.Y -= float64(player.Speed) * 0.75
		player.X += float64(player.Speed) * 0.75
	} else if (ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)) && (ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)) {
		player.Y += float64(player.Speed) * 0.75
		player.X -= float64(player.Speed) * 0.75
	} else if (ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)) && (ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)) {
		player.Y += float64(player.Speed) * 0.75
		player.X += float64(player.Speed) * 0.75
	} else {
		if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			player.X += float64(player.Speed)
		}
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			player.X -= float64(player.Speed)
		}
		if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
			player.Y -= float64(player.Speed)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
			player.Y += float64(player.Speed)
		}
	}

}

func (g *Game) followCursor() {
	cursorX, cursorY := ebiten.CursorPosition()
	windowX, windowY := ebiten.WindowSize()
	if cursorX > windowX || cursorX < 0 || cursorY > windowY || cursorY < 0 {
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
			X:     (rand.Float64() * 320),
			Y:     (rand.Float64() * 240),
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
