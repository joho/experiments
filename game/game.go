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

	fullScene  *sprite.Node
	background *sprite.Node
	foreground *sprite.Node
	eng        = glsprite.Engine()

	shipDeltaV   float32 = 1.2
	bulletDeltaV geom.Pt = 3

	shipPos = geom.Point{100, 100}

	nextShipPos *geom.Point

	bottomRight *geom.Point
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
	currentBottomRight := geom.Point{c.Width, c.Height}
	if bottomRight == nil || currentBottomRight != *bottomRight {
		bottomRight = &currentBottomRight

		log.Printf("Device Sizing: %vx%v PixelsPerPt:%v",
			c.Width,
			c.Height,
			c.PixelsPerPt,
		)
	}

	if fullScene == nil {
		fullScene = setupScene()
	}

	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	now := clock.Time(time.Since(startTime) * 60 / time.Second)

	eng.Render(fullScene, now, c)
	debug.DrawFPS(c)
}

func touch(t event.Touch, c event.Config) {
	log.Printf("touch at %v:%v\n", t.Loc.X, t.Loc.Y)

	bottomTenthY := c.Height - c.Height/8
	if t.Loc.Y > bottomTenthY {
		log.Println("FIRE ZE MISSLES")

		firingPoint := shipPos

		bullet := loadSprite("bullet.png")
		bulletNode := newNode(background)
		bulletNode.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
			eng.SetSubTex(n, bullet.SubTex)

			bulletScalingFactor := 4
			width := float32(bullet.Width / bulletScalingFactor)
			height := float32(bullet.Height / bulletScalingFactor)

			x := float32(firingPoint.X) - float32(bullet.Width/(2*bulletScalingFactor))
			y := float32(firingPoint.Y) - 20 // magic number for tip of ship

			eng.SetTransform(n, f32.Affine{
				{width, 0, x},
				{0, height, y},
			})

			firingPoint.Y = firingPoint.Y - bulletDeltaV
		})

	} else {
		nextShipPos = &t.Loc
	}
}

func setupScene() *sprite.Node {
	fullScene = &sprite.Node{}
	eng.Register(fullScene)
	eng.SetTransform(fullScene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	background = newNode(fullScene)
	foreground = newNode(fullScene)

	playerShip := loadSprite("player_ship.png")
	shipNode := newNode(foreground)

	shipNode.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, playerShip.SubTex)

		width := float32(playerShip.Width)
		height := float32(playerShip.Height)

		if nextShipPos != nil {
			if *nextShipPos == shipPos {
				nextShipPos = nil
			} else {
				if nextShipPos.X > shipPos.X {
					shipPos.X = geom.Pt(float32(shipPos.X) + shipDeltaV)
				}
				if nextShipPos.X < shipPos.X {
					shipPos.X = geom.Pt(float32(shipPos.X) - shipDeltaV)
				}
				if nextShipPos.Y > shipPos.Y {
					shipPos.Y = geom.Pt(float32(shipPos.Y) + shipDeltaV)
				}
				if nextShipPos.Y < shipPos.Y {
					shipPos.Y = geom.Pt(float32(shipPos.Y) - shipDeltaV)
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

	return fullScene
}

func newNode(scene *sprite.Node) *sprite.Node {
	node := &sprite.Node{}
	eng.Register(node)
	scene.AppendChild(node)
	return node
}

type Sprite struct {
	sprite.SubTex
	Width, Height int
}

func loadSprite(fileName string) Sprite {
	a, err := asset.Open(fileName)
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
