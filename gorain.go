package main

import (
	"flag"
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
	lightningChance       = 0.005
	lightningGrowthDelay  = 2 * time.Millisecond
	lightningMaxBranches  = 2
	lightningBranchChance = 0.3
	forkChance            = 0.15
	forkHorizontalSpread  = 3
	segmentLifespan       = 800 * time.Millisecond
)

// Raindrop represents a falling raindrop
type Raindrop struct {
	x, y  float64
	speed float64
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
	lightningChars = []rune{'*', '+', '#'} // Dimmest to brightest
	rainStyle      tcell.Style
	lightningStyle tcell.Style
)

// TODO: update here
// wdym update here? like adding wat??

// setupColors initializes styles for rain and lightning
func setupColors(rainColor, lightningColor string) {
	fgRain, ok := colorMap[rainColor]
	if !ok {
		fgRain = tcell.ColorAqua
	}
	fgLightning, ok := colorMap[lightningColor]
	if !ok {
		fgLightning = tcell.ColorYellow
	}
	rainStyle = tcell.StyleDefault.Foreground(fgRain).Background(tcell.ColorDefault)
	lightningStyle = tcell.StyleDefault.Foreground(fgLightning).Background(tcell.ColorDefault).Bold(true)
}

// updateBolt updates the lightning bolt and returns true if it should continue existing
func (bolt *LightningBolt) update() bool {
	currentTime := time.Now()

	// growth
	if bolt.isGrowing && currentTime.Sub(bolt.lastGrowthTime) >= lightningGrowthDelay {
		bolt.lastGrowthTime = currentTime
		var newSegments []LightningSegment
		addedSegment := false
		lastSegment := bolt.segments[len(bolt.segments)-1]
		lastY, lastX := lastSegment.y, lastSegment.x

		if len(bolt.segments) < bolt.targetLength && lastY < bolt.maxY-1 {
			branches := 1
			if rand.Float64() < lightningBranchChance {
				branches = rand.Intn(lightningMaxBranches+1) + 1
			}
			currentX := lastX
			var nextPrimaryX int
			for i := 0; i < branches; i++ {
				offset := rand.Intn(5) - 2 // -2 to 2
				nextX := max(0, min(bolt.maxX-1, currentX+offset))
				nextY := min(bolt.maxY-1, lastY+1)
				newSegments = append(newSegments, LightningSegment{nextY, nextX, currentTime})
				if i == 0 {
					nextPrimaryX = nextX
				}
				currentX = nextX
				addedSegment = true
			}

			// add secondary forks
			if rand.Float64() < forkChance {
				forkOffset := rand.Intn(2*forkHorizontalSpread+1) - forkHorizontalSpread
				if forkOffset == 0 {
					if rand.Intn(2) == 0 {
						forkOffset = -1
					} else {
						forkOffset = 1
					}
				}
				forkX := max(0, min(bolt.maxX-1, lastX+forkOffset))
				forkY := min(bolt.maxY-1, lastY+1)
				if forkX != nextPrimaryX {
					newSegments = append(newSegments, LightningSegment{forkY, forkX, currentTime})
					addedSegment = true
				}
			}

			if !addedSegment || len(bolt.segments) >= bolt.targetLength || lastY >= bolt.maxY-1 {
				bolt.isGrowing = false
			}

			bolt.segments = append(bolt.segments, newSegments...)
		}
	}

	// check for removal
	allExpired := true
	for _, seg := range bolt.segments {
		if currentTime.Sub(seg.createdTime) <= segmentLifespan {
			allExpired = false
			break
		}
	}
	return !allExpired
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
				charIndex = 2 // '#'
			} else if normAge < 0.66 {
				charIndex = 1 // '+'
			} else {
				charIndex = 0 // '*'
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
func simulateRain(screen tcell.Screen, rainColor, lightningColor string) error {
	setupColors(rainColor, lightningColor)
	rand.Seed(time.Now().UnixNano())
	raindrops := []Raindrop{}
	activeBolts := []*LightningBolt{}
	isThunderstorm := false

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
					if ev.Rune() == 'q' || ev.Rune() == 'Q' {
						return nil
					}
					if ev.Rune() == 't' || ev.Rune() == 'T' {
						isThunderstorm = !isThunderstorm
						screen.Clear()
					}
				case tcell.KeyEscape:
					return nil
				case tcell.KeyCtrlC:
					return nil
				}
			case *tcell.EventResize:
				screen.Clear()
				raindrops = nil
				activeBolts = nil
			}
		default:
			// frame rate control
			currentTime := time.Now()
			deltaTime := currentTime.Sub(lastUpdateTime)
			if deltaTime < updateInterval {
				time.Sleep(updateInterval - deltaTime)
			}
			lastUpdateTime = time.Now()

			// update lightning
			maxX, maxY := screen.Size()
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
			generationChance := 0.5
			maxNewDrops := maxX / 8
			minSpeed, maxSpeed := 0.3, 1.0
			if !isThunderstorm {
				generationChance = 0.3
				maxNewDrops = maxX / 15
				maxSpeed = 0.6
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
			var nextRaindrops []Raindrop
			for _, drop := range raindrops {
				drop.y += drop.speed
				if int(drop.y) < maxY {
					nextRaindrops = append(nextRaindrops, drop)
				}
			}
			raindrops = nextRaindrops

			// Draw
			screen.Clear()
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
			screen.Show()
		}
	}
}

func main() {
	// command line flags
	rainColor := flag.String("rain-color", "cyan", "Color for the rain (black, red, green, yellow, blue, magenta, cyan, white)")
	lightningColor := flag.String("lightning-color", "yellow", "Color for the lightning (black, red, green, yellow, blue, magenta, cyan, white)")
	flag.Parse()

	// validate terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		println("Error: This program requires a TTY with terminal support.")
		os.Exit(1)
	}

	// initialize tcell
	screen, err := tcell.NewScreen()
	if err != nil {
		println("Error initializing screen:", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		println("Error initializing screen:", err)
		os.Exit(1)
	}
	defer screen.Fini()

	// run simulation
	if err := simulateRain(screen, *rainColor, *lightningColor); err != nil {
		screen.Fini()
		println("Error during simulation:", err)
		os.Exit(1)
	}
}
