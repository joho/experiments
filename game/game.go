package main

import (
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
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

var (
	startTime = time.Now()

	scene *sprite.Node
	eng   = glsprite.Engine()

	deltaV float32 = 1.2

	shipPos = geom.Point{100, 100}

	nextShipPos *geom.Point
)

func main() {
	app.Run(app.Callbacks{
		Start: start,
		Draw:  draw,
		Stop:  stop,
		Touch: touch,
	})
}

func start() {
	log.Println("starting app")
}

func stop() {
	log.Println("stopping app")
}

func draw(c event.Config) {
	if scene == nil {
		scene = setupScene()

		log.Printf("Device Stats: Size:%vx%v PixelsPerPt:%v",
			c.Width,
			c.Height,
			c.PixelsPerPt,
		)
	}

	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	now := clock.Time(time.Since(startTime) * 60 / time.Second)

	eng.Render(scene, now, c)
	debug.DrawFPS(c)
}

func touch(t event.Touch, c event.Config) {
	nextShipPos = &t.Loc
	log.Printf("touch at %v:%v\n", t.Loc.X, t.Loc.Y)
}

func setupScene() *sprite.Node {
	scene := &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	playerShip := loadSprite("player_ship.png")

	shipNode := &sprite.Node{}
	eng.Register(shipNode)
	scene.AppendChild(shipNode)
	shipNode.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, playerShip.SubTex)

		width := float32(playerShip.Width)
		height := float32(playerShip.Height)

		if nextShipPos != nil {
			if *nextShipPos == shipPos {
				nextShipPos = nil
			} else {
				if nextShipPos.X > shipPos.X {
					shipPos.X = geom.Pt(float32(shipPos.X) + deltaV)
				}
				if nextShipPos.X < shipPos.X {
					shipPos.X = geom.Pt(float32(shipPos.X) - deltaV)
				}
				if nextShipPos.Y > shipPos.Y {
					shipPos.Y = geom.Pt(float32(shipPos.Y) + deltaV)
				}
				if nextShipPos.Y < shipPos.Y {
					shipPos.Y = geom.Pt(float32(shipPos.Y) - deltaV)
				}
			}
		}

		x := float32(shipPos.X) - width/2
		y := float32(shipPos.Y) - height/2

		eng.SetTransform(n, f32.Affine{
			{width, 0, x},
			{0, height, y},
		})
	})

	return scene
}

type Sprite struct {
	sprite.SubTex
	Width, Height int
}

func loadSprite(fileName string) Sprite {
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

	imgWidth := bounds.Max.X - bounds.Min.X
	imgHeight := bounds.Max.Y - bounds.Min.Y
	log.Printf("sprite %v size: %vx%v\n", fileName, imgWidth, imgHeight)

	subTex := sprite.SubTex{t, image.Rect(0, 0, imgWidth, imgHeight)}

	return Sprite{
		SubTex: subTex,
		Width:  imgWidth,
		Height: imgHeight,
	}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
