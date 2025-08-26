package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"
)

// configuration constants
const (
	updateInterval        = 15 * time.Millisecond
	rainChars             = "|.`"
	snowChars             = "*-"
	lightningChance       = 0.005
	lightningGrowthDelay  = 2 * time.Millisecond
	lightningMaxBranches  = 2
	lightningBranchChance = 0.3
	forkChance            = 0.15
	forkHorizontalSpread  = 3
	segmentLifespan       = 800 * time.Millisecond
	starCount             = 50
	moonRadius            = 3
	moonYPosition         = 5
)

// Raindrop represents a falling raindrop
type Raindrop struct {
	x, y  float64
	speed float64
	char  rune
}

// Snowflake represents a falling snowflake
type Snowflake struct {
	x, y  float64
	speed float64
	drift float64
	char  rune
}

// LightningBolt represents a lightning bolt with segments
type LightningBolt struct {
	segments       []LightningSegment
	startCol       int
	targetLength   int
	lastGrowthTime time.Time
	isGrowing      bool
	maxY, maxX     int
}

// LightningSegment represents a single segment of a lightning bolt
type LightningSegment struct {
	y, x        int
	createdTime time.Time
}

// Star represents a background star
type Star struct {
	x, y float64
	char rune
}

// color mapping
var colorMap = map[string]tcell.Color{
	"black":   tcell.ColorBlack,
	"red":     tcell.ColorRed,
	"green":   tcell.ColorGreen,
	"yellow":  tcell.ColorYellow,
	"blue":    tcell.ColorBlue,
	"magenta": tcell.ColorPurple,
	"cyan":    tcell.ColorAqua,
	"white":   tcell.ColorWhite,
}

var (
	lightningChars = []rune{'*', '+', '#'}
	rainStyle      tcell.Style
	lightningStyle tcell.Style
	snowStyle      tcell.Style
	//starStyle      tcell.Style
	moonStyle tcell.Style
	moonChar  rune
)

// setupColors initializes styles for all elements
func setupColors(rainColor, lightningColor, snowColor, moonColor, mChar string) {
	fgRain, ok := colorMap[rainColor]
	if !ok {
		fgRain = tcell.ColorAqua
	}
	fgLightning, ok := colorMap[lightningColor]
	if !ok {
		fgLightning = tcell.ColorYellow
	}
	fgSnow, ok := colorMap[snowColor]
	if !ok {
		fgSnow = tcell.ColorWhite
	}
	fgMoon, ok := colorMap[moonColor]
	if !ok {
		fgMoon = tcell.ColorYellow
	}
	moonChar = '#'
	if len(mChar) > 0 {
		moonChar = []rune(mChar)[0]
	}

	rainStyle = tcell.StyleDefault.Foreground(fgRain).Background(tcell.ColorDefault)
	lightningStyle = tcell.StyleDefault.Foreground(fgLightning).Background(tcell.ColorDefault).Bold(true)
	snowStyle = tcell.StyleDefault.Foreground(fgSnow).Background(tcell.ColorDefault)
	//starStyle = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDefault).Dim(true)
	moonStyle = tcell.StyleDefault.Foreground(fgMoon).Background(tcell.ColorDefault).Bold(true)
}

// drawMoon draws an ASCII art moon circle
func drawMoon(screen tcell.Screen, centerX int, char rune, style tcell.Style) {
	for y := -moonRadius; y <= moonRadius; y++ {
		for x := -moonRadius * 2; x <= moonRadius*2; x++ {
			distance := math.Sqrt(float64(x*x)/4 + float64(y*y))
			if distance <= float64(moonRadius)+0.5 {
				screen.SetContent(centerX+x, moonYPosition+y, char, nil, style)
			}
		}
	}
}

// drawBolt draws the lightning bolt based on segment age
func (bolt *LightningBolt) draw(screen tcell.Screen) {
	currentTime := time.Now()
	maxCharIndex := len(lightningChars) - 1

	for _, seg := range bolt.segments {
		segmentAge := currentTime.Sub(seg.createdTime)
		var char rune
		isVisible := true

		if segmentAge <= segmentLifespan {
			normAge := float64(segmentAge) / float64(segmentLifespan)
			charIndex := 0
			if normAge < 0.33 {
				charIndex = 2
			} else if normAge < 0.66 {
				charIndex = 1
			} else {
				charIndex = 0
			}
			charIndex = max(0, min(maxCharIndex, charIndex))
			char = lightningChars[charIndex]
		} else {
			isVisible = false
		}

		if isVisible {
			_, maxX := screen.Size()
			if seg.y < bolt.maxY && seg.x < maxX {
				screen.SetContent(seg.x, seg.y, char, nil, lightningStyle)
			}
		}
	}
}

// updateBolt updates the lightning bolt
func (bolt *LightningBolt) update() bool {
	currentTime := time.Now()
	if bolt.isGrowing && currentTime.Sub(bolt.lastGrowthTime) >= lightningGrowthDelay {
		bolt.lastGrowthTime = currentTime
		var newSegments []LightningSegment
		lastSegment := bolt.segments[len(bolt.segments)-1]
		lastY, lastX := lastSegment.y, lastSegment.x
		if len(bolt.segments) < bolt.targetLength && lastY < bolt.maxY-1 {
			branches := 1
			if rand.Float64() < lightningBranchChance {
				branches = rand.Intn(lightningMaxBranches+1) + 1
			}
			currentX := lastX
			for i := 0; i < branches; i++ {
				offset := rand.Intn(5) - 2
				nextX := max(0, min(bolt.maxX-1, currentX+offset))
				nextY := min(bolt.maxY-1, lastY+1)
				newSegments = append(newSegments, LightningSegment{nextY, nextX, currentTime})
				currentX = nextX
			}
			if len(bolt.segments) >= bolt.targetLength || lastY >= bolt.maxY-1 {
				bolt.isGrowing = false
			}
			bolt.segments = append(bolt.segments, newSegments...)
		}
	}
	allExpired := true
	for _, seg := range bolt.segments {
		if currentTime.Sub(seg.createdTime) <= segmentLifespan {
			allExpired = false
			break
		}
	}
	return !allExpired
}

// helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// simulateRain runs the main simulation loop
func simulateRain(screen tcell.Screen, rainColor, lightningColor, snowColor, moonColor, moonCharacter string) error {
	setupColors(rainColor, lightningColor, snowColor, moonColor, moonCharacter)
	rand.Seed(time.Now().UnixNano())
	raindrops := []Raindrop{}
	snowflakes := []Snowflake{}
	activeBolts := []*LightningBolt{}
	stars := []Star{}

	isThunderstorm := false
	isSnowing := false
	isNight := false
	isWeatherHidden := false

	// event channel for input
	events := make(chan tcell.Event)
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	lastUpdateTime := time.Now()
	for {
		// input handler
		select {
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyRune:
					switch ev.Rune() {
					case 'q', 'Q':
						return nil
					case 't', 'T':
						isThunderstorm = !isThunderstorm
						isSnowing = false
						screen.Clear()
					case 's', 'S':
						isSnowing = !isSnowing
						isThunderstorm = false
						screen.Clear()
					case 'n', 'N':
						isNight = !isNight
						screen.Clear()
						if isNight {
							maxX, maxY := screen.Size()
							stars = nil
							for i := 0; i < starCount; i++ {
								x := float64(rand.Intn(maxX))
								y := float64(rand.Intn(maxY / 2))
								stars = append(stars, Star{x: x, y: y, char: '•'})
							}
						}
					case 'h', 'H':
						isWeatherHidden = !isWeatherHidden
						screen.Clear()
					}
				case tcell.KeyEscape, tcell.KeyCtrlC:
					return nil
				}
			case *tcell.EventResize:
				screen.Clear()
				raindrops = nil
				snowflakes = nil
				activeBolts = nil
				stars = nil
				if isNight {
					maxX, maxY := screen.Size()
					for i := 0; i < starCount; i++ {
						x := float64(rand.Intn(maxX))
						y := float64(rand.Intn(maxY / 2))
						stars = append(stars, Star{x: x, y: y, char: '.'})
					}
				}
			}
		default:
			// frame rate control
			currentTime := time.Now()
			deltaTime := currentTime.Sub(lastUpdateTime)
			if deltaTime < updateInterval {
				time.Sleep(updateInterval - deltaTime)
			}
			lastUpdateTime = time.Now()

			maxX, maxY := screen.Size()

			// Update lightning
			var nextBolts []*LightningBolt
			if isThunderstorm && len(activeBolts) < 3 && rand.Float64() < lightningChance {
				startCol := rand.Intn(maxX/2) + maxX/4
				startRow := rand.Intn(maxY / 5)
				targetLength := rand.Intn(maxY-maxY/2-2) + maxY/2
				activeBolts = append(activeBolts, &LightningBolt{
					segments:       []LightningSegment{{startRow, startCol, time.Now()}},
					startCol:       startCol,
					targetLength:   targetLength,
					lastGrowthTime: time.Now(),
					isGrowing:      true,
					maxY:           maxY,
					maxX:           maxX,
				})
			}
			for _, bolt := range activeBolts {
				if bolt.update() {
					nextBolts = append(nextBolts, bolt)
				}
			}
			activeBolts = nextBolts

			// Update raindrops
			if !isWeatherHidden && !isSnowing {
				generationChance := 0.3
				maxNewDrops := maxX / 15
				minSpeed, maxSpeed := 0.3, 0.6
				if isThunderstorm {
					generationChance = 0.5
					maxNewDrops = maxX / 8
					maxSpeed = 1.0
				}
				if rand.Float64() < generationChance {
					numNewDrops := rand.Intn(maxNewDrops) + 1
					for i := 0; i < numNewDrops; i++ {
						x := rand.Intn(maxX)
						speed := rand.Float64()*(maxSpeed-minSpeed) + minSpeed
						char := rainChars[rand.Intn(len(rainChars))]
						raindrops = append(raindrops, Raindrop{x: float64(x), y: 0, speed: speed, char: rune(char)})
					}
				}
			}
			var nextRaindrops []Raindrop
			for _, drop := range raindrops {
				drop.y += drop.speed
				if int(drop.y) < maxY {
					nextRaindrops = append(nextRaindrops, drop)
				}
			}
			raindrops = nextRaindrops

			// Update snowflakes
			if !isWeatherHidden && isSnowing {
				snowGenChance := 0.2
				maxNewFlakes := maxX / 10
				snowMinSpeed, snowMaxSpeed := 0.05, 0.2
				if rand.Float64() < snowGenChance {
					numNewFlakes := rand.Intn(maxNewFlakes) + 1
					for i := 0; i < numNewFlakes; i++ {
						x := rand.Intn(maxX)
						speed := rand.Float64()*(snowMaxSpeed-snowMinSpeed) + snowMinSpeed
						drift := (rand.Float64() - 0.5) * 0.2
						char := snowChars[rand.Intn(len(snowChars))]
						snowflakes = append(snowflakes, Snowflake{x: float64(x), y: 0, speed: speed, drift: drift, char: rune(char)})
					}
				}
			}
			var nextSnowflakes []Snowflake
			for _, flake := range snowflakes {
				flake.y += flake.speed
				flake.x += flake.drift
				if int(flake.y) < maxY && int(flake.x) >= 0 && int(flake.x) < maxX {
					nextSnowflakes = append(nextSnowflakes, flake)
				}
			}
			snowflakes = nextSnowflakes

			// drawing
			screen.Clear()

			// draw background if night mode is active (always draw first)
			if isNight {
				for _, star := range stars {
					if int(star.x) < maxX && int(star.y) < maxY {
						// set the star's base style. The color is now conditional based on the weather.
						style := tcell.StyleDefault.Background(tcell.ColorDefault).Dim(true)
						if isSnowing {
							style = style.Foreground(tcell.ColorGray)
						} else {
							style = style.Foreground(tcell.ColorWhite)
						}

						// add a small random chance to make the star appear bold (brighter)
						if rand.Float64() < 0.005 {
							style = style.Bold(true)
						}
						screen.SetContent(int(star.x), int(star.y), star.char, nil, style)
					}
				}
				drawMoon(screen, maxX/2, moonChar, moonStyle)
			}

			// draw weather effects if not hidden (always draw on top of background)
			if !isWeatherHidden {
				for _, bolt := range activeBolts {
					bolt.draw(screen)
				}
				for _, drop := range raindrops {
					if int(drop.y) < maxY {
						style := rainStyle
						if isThunderstorm {
							style = style.Bold(true)
						} else if drop.speed < 0.8 {
							style = style.Dim(true)
						}
						screen.SetContent(int(drop.x), int(drop.y), drop.char, nil, style)
					}
				}
				for _, flake := range snowflakes {
					if int(flake.y) < maxY && int(flake.x) >= 0 && int(flake.x) < maxX {
						screen.SetContent(int(flake.x), int(flake.y), flake.char, nil, snowStyle)
					}
				}
			}
			screen.Show()
		}
	}
}

func main() {
	// command line flags
	rainColor := flag.String("rain-color", "cyan", "Color for the rain")
	lightningColor := flag.String("lightning-color", "yellow", "Color for the lightning")
	snowColor := flag.String("snow-color", "white", "Color for the snow")
	moonColor := flag.String("moon-color", "yellow", "Color for the moon")
	moonCharacter := flag.String("moon-char", "#", "Character for the moon")
	flag.Parse()

	// validate terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println("Error: This program requires a TTY with terminal support.")
		os.Exit(1)
	}

	// initialize tcell
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Println("Error initializing screen:", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Println("Error initializing screen:", err)
		os.Exit(1)
	}
	defer screen.Fini()

	// run simulation
	if err := simulateRain(screen, *rainColor, *lightningColor, *snowColor, *moonColor, *moonCharacter); err != nil {
		screen.Fini()
		fmt.Println("Error during simulation:", err)
		os.Exit(1)
	}
}

// parrot :P
