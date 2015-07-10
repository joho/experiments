package main

import (
	"fmt"
	"image"
	"log"
	"time"

	_ "image/png"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

const (
	width  = 640
	height = 480
)

var (
	startTime = time.Now()

	scene *sprite.Node
	eng   = glsprite.Engine()
)

func main() {
	app.Run(app.Callbacks{
		Start: start,
		Draw:  draw,
		Stop:  stop,
	})
}

func start() {
	fmt.Println("starting app")
}

func stop() {
	fmt.Println("stopping app")
}

func draw(c event.Config) {
	if scene == nil {
		scene = setupScene()
	}

	gl.ClearColor(0, 0, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	now := clock.Time(time.Since(startTime) * 60 / time.Second)

	eng.Render(scene, now, c)
	debug.DrawFPS(c)
}

func setupScene() *sprite.Node {
	scene := &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	playerShip := loadPlayerShip()

	shipNode := &sprite.Node{}
	eng.Register(shipNode)
	scene.AppendChild(shipNode)
	shipNode.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, playerShip)

		eng.SetTransform(n, f32.Affine{
			{width, 0, 100},
			{0, height, 100},
		})
	})

	return scene
}

func loadPlayerShip() sprite.SubTex {
	a, err := asset.Open("player_ship.png")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	img, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}
	t, err := eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()

	shipWidth := bounds.Max.X - bounds.Min.X
	shipHeight := bounds.Max.Y - bounds.Min.Y

	return sprite.SubTex{t, image.Rect(0, 0, shipWidth, shipHeight)}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
