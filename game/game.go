package main

import (
	"image"
	"log"
	"math/rand"
	"time"

	_ "image/png"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

const (
	shipDeltaV   float32 = 1.2
	enemyDeltaV  float32 = 0.7
	bulletDeltaV geom.Pt = 3
)

var (
	startTime = time.Now()

	fullScene  *sprite.Node
	background *sprite.Node
	foreground *sprite.Node
	eng        = glsprite.Engine()

	shipPos *geom.Point

	nextShipPos *geom.Point

	bottomRight *geom.Point

	bullets = []Bullet{}

	enemySprite *Sprite
	enemies     = []Enemy{}

	screenWidth *geom.Pt
)

func main() {
	var conf config.Event
	app.Main(func(a app.App) {
		for e := range a.Events() {
			switch ee := app.Filter(e).(type) {
			case paint.Event:
				draw(conf)
				a.EndPaint()
			case touch.Event:
				onTouch(ee, conf)
			case config.Event:
				conf = ee
			case lifecycle.Event:
				switch ee.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					start()
				case lifecycle.CrossOff:
					// occasionally doesn't work and need to CTRL+C the console
					stop()
					return
				}
			}
		}
	})
}

func start() {
	log.Println("starting app")
}

func stop() {
	log.Println("stopping app")
}

func draw(c config.Event) {
	secondsFromStart := time.Since(startTime) * 60 / time.Second
	now := clock.Time(secondsFromStart)

	if screenWidth == nil || *screenWidth != c.Width {
		screenWidth = &c.Width

		log.Printf("Device Sizing: %vx%v PixelsPerPt:%v",
			c.Width,
			c.Height,
			c.PixelsPerPt,
		)
	}

	if fullScene == nil {
		fullScene = setupScene(c.Width, c.Height, secondsFromStart)
	}

	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	eng.Render(fullScene, now, c)
	debug.DrawFPS(c)
}

func onTouch(t touch.Event, c config.Event) {
	log.Printf("touch at %v:%v\n", t.Loc.X, t.Loc.Y)

	bottomTenthY := c.Height - c.Height/8
	if t.Loc.Y > bottomTenthY {
		fireBullet(t.Loc)
	} else {
		setShipDestination(t.Loc)
	}
}

func setShipDestination(loc geom.Point) {
	nextShipPos = &loc
}

func fireBullet(loc geom.Point) {
	log.Println("FIRE ZE MISSLES")

	firingPoint := *shipPos

	bullet := loadSprite("bullet.png")
	bulletNode := newNode(background)
	bulletNode.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, bullet.SubTex)

		const bulletScalingFactor = 4
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
	bullets = append(bullets, Bullet{bullet})
}

func setupScene(width, height geom.Pt, secondsFromStart time.Duration) *sprite.Node {
	fullScene = &sprite.Node{}
	eng.Register(fullScene)
	eng.SetTransform(fullScene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	background = newNode(fullScene)
	foreground = newNode(fullScene)

	setupShip(foreground, width)

	go func() {
		for {
			setupEnemy(foreground, *screenWidth)
			time.Sleep(2 * time.Second)
		}
	}()

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

type Bullet struct {
	Sprite
}

type Enemy struct {
	Sprite
}

func setupShip(parentNode *sprite.Node, screenWidth geom.Pt) {
	playerShip := loadSprite("player_ship.png")
	shipNode := newNode(foreground)
	if shipPos == nil {
		shipPos = &geom.Point{
			screenWidth / 2,
			100,
		}
	}

	shipNode.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, playerShip.SubTex)

		width := float32(playerShip.Width)
		height := float32(playerShip.Height)

		if nextShipPos != nil {
			if *nextShipPos == *shipPos {
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
}

func setupEnemy(parentNode *sprite.Node, screenWidth geom.Pt) {
	if enemySprite == nil {
		sprite := loadSprite("baddy_one.png")
		enemySprite = &sprite
	}

	enemy := Enemy{*enemySprite}
	enemyNode := newNode(parentNode)

	const enemyScalingFactor = 4
	width := float32(enemy.Width / enemyScalingFactor)
	height := float32(enemy.Height / enemyScalingFactor)

	x := float32(rand.Intn(int(screenWidth) - int(width)))
	var y float32 = 20

	enemyNode.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, enemy.SubTex)

		eng.SetTransform(n, f32.Affine{
			{width, 0, x},
			{0, height, y},
		})

		y = y + enemyDeltaV
	})

	enemies = append(enemies, enemy)
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
