package main

import (
	"fmt"
	"image"
	"log"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/danelee/rpg-test/src/entities"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// const
const SCREEN_WIDTH int = 640
const SCREEN_HEIGHT int = 480

// Game implements ebiten.Game interface.
type Game struct {
	player        *entities.Player
	enemies       []*entities.Enemy
	camera        *Camera
	tilemap       *Tilemap
	tilemapImage  *ebiten.Image
	projectile    []*entities.Projectile
	canSpawnEnemy bool
	spawnTimer    time.Time
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write game's logic update.

	g.movePlayer()

	g.camera.Follow(g.player.X+8, g.player.Y+8, float64(SCREEN_WIDTH), float64(SCREEN_HEIGHT))
	g.camera.Constrain(
		float64(g.tilemap.Layers[0].Width)*16.0,
		float64(g.tilemap.Layers[0].Height)*16.0,
		float64(SCREEN_WIDTH),
		float64(SCREEN_HEIGHT))

	for _, enemy := range g.enemies {
		g.followPlayer(enemy)
		g.attackEnemy(enemy)
	}
	if g.canSpawnEnemy {
		g.spawnEnemy()
		g.canSpawnEnemy = false
		//g.spawnTimer = time.Now()
	}
	g.checkCanSpawnEnemy()

	//if ebiten.CursorMode() == 0 {
	//	g.followCursor()
	//}

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {

	// Write your game's rendering.
	//screen.Fill(color.RGBA{192, 192, 192, 127})

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
		//ebitenutil.DebugPrintAt(screen, fmt.Sprintf("enemy X = %v, enemy Y = %v, enemy Health = %v", enemy.X, enemy.Y, enemy.Health), 5, 300)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS = %v, TPS = %v", ebiten.ActualFPS(), ebiten.ActualTPS()))

	//draw projectiles
	opts.GeoM.Reset()
	if len(g.projectile) != 0 && g.projectile[0].Enabled {
		g.drawProjectile(screen, &opts, g.projectile[0])
		//ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Pojectile drawn. X = %v, Y = %v", g.projectile[0].X, g.projectile[0].Y), 5, 320)
	}

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
		player:        player,
		enemies:       []*entities.Enemy{},
		camera:        NewCamera(0, 0),
		tilemap:       tilemap,
		tilemapImage:  tilemapImage,
		projectile:    []*entities.Projectile{},
		canSpawnEnemy: true,
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
		Speed:   1,
		Sprite: &entities.Sprite{
			X:     (rand.Float64() * float64(SCREEN_WIDTH)),
			Y:     (rand.Float64() * float64(SCREEN_HEIGHT)),
			Image: image,
		},
	}

	return enemy, nil
}

// make generic so enemy follows player, projectile follows enemy
func (g *Game) followPlayer(enemy *entities.Enemy) {
	offSet := 500.00

	boxMinX, boxMaxX, boxMinY, boxMaxY := enemy.X-offSet, enemy.X+offSet, enemy.Y-offSet, enemy.Y+offSet

	//if player is within a certain box distance of the enemy
	inBox := g.player.X >= boxMinX && g.player.X <= boxMaxX && g.player.Y >= boxMinY && g.player.Y <= boxMaxY

	if inBox {
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
}

func (g *Game) attackEnemy(enemy *entities.Enemy) {
	offSet := 50.00

	boxMinX, boxMaxX, boxMinY, boxMaxY := g.player.X-offSet, g.player.X+offSet, g.player.Y-offSet, g.player.Y+offSet

	inBox := enemy.X >= boxMinX && enemy.X <= boxMaxX && enemy.Y >= boxMinY && enemy.Y <= boxMaxY

	if inBox && len(g.projectile) == 0 {

		proj := entities.Projectile{
			Sprite: &entities.Sprite{
				X:     g.player.X,
				Y:     g.player.Y,
				Image: g.player.Image,
			},
			Damage:  g.player.Attack,
			Speed:   g.player.Speed * 2,
			Enabled: true,
		}

		g.projectile = append(g.projectile, &proj)
		//player shoots
	}
	if len(g.projectile) > 0 {
		if g.projectile[0].X < enemy.X {
			g.projectile[0].X += float64(g.projectile[0].Speed)
		}
		if g.projectile[0].Y < enemy.Y {
			g.projectile[0].Y += float64(g.projectile[0].Speed)
		}
		if g.projectile[0].X > enemy.X {
			g.projectile[0].X -= float64(g.projectile[0].Speed)
		}
		if g.projectile[0].Y > enemy.Y {
			g.projectile[0].Y -= float64(g.projectile[0].Speed)
		}

		eX1, eY1, eX2, eY2 := enemy.X, enemy.Y, enemy.X+16, enemy.Y+16
		pX1, pY1, pX2, pY2 := g.projectile[0].X, g.projectile[0].Y, g.projectile[0].X+12, g.projectile[0].Y+12

		projHit := (pX2 >= eX1 && pX2 <= eX2 && pY2 >= eY1 && pY2 <= eY2) || (pX1 >= eX1 && pX1 <= eX2 && pY1 >= eY1 && pY1 <= eY2) || (pX1 >= eX1 && pX1 <= eX2 && pY2 >= eY1 && pY2 <= eY2) || (pX2 >= eX1 && pX2 <= eX2 && pY1 >= eY1 && pY1 <= eY2)
		if projHit {
			enemy.Health -= g.player.Attack
			g.projectile[0].Enabled = false
			g.projectile = slices.DeleteFunc(g.projectile, func(proj *entities.Projectile) bool {
				return !proj.Enabled
			})
		}
	}

	if enemy.Health == 0 {
		g.enemies = slices.DeleteFunc(g.enemies, func(enemy *entities.Enemy) bool {
			return enemy.Health == 0
		})
	}
}

func (g *Game) drawProjectile(screen *ebiten.Image, opts *ebiten.DrawImageOptions, projectile *entities.Projectile) {
	opts.GeoM.Translate(projectile.X, projectile.Y)
	opts.GeoM.Translate(g.camera.X, g.camera.Y)
	screen.DrawImage(g.player.Image.SubImage(image.Rect(2, 2, 14, 14)).(*ebiten.Image), opts)
	opts.GeoM.Reset()
}

func (g *Game) spawnEnemy() {

	enemy, err := loadEnemy()
	if err != nil {
		log.Fatal(err)
	}

	g.enemies = append(g.enemies, enemy)

}

func (g *Game) checkCanSpawnEnemy() {
	if !g.canSpawnEnemy {
		if len(g.enemies) == 0 {
			g.canSpawnEnemy = true
		}
	}
}

//TODO add projectile that shoots the enemy
//Implement projectile damage and enemy death
//Add multiple enemies with different spawn rate
