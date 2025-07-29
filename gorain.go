package main

import (
	"flag"
	"fmt"

	//"io" will use later (probably)
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/gdamore/tcell/v2"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

// Configuration constants
const (
	updateInterval        = 15 * time.Millisecond
	rainChars             = "|.`"
	lightningChance       = 0.1 // Increased for easier testing, adjust as needed
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

// Color mapping
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

// initAudio initializes the audio speaker once.
// It opens a sound file to get the format for speaker initialization.
func initAudio() error {
	f, err := os.Open("assets/rain-sound.wav")
	if err != nil {
		return fmt.Errorf("error opening assets/rain-sound.wav for speaker initialization: %w", err)
	}
	defer f.Close()

	streamer, format, err := wav.Decode(f)
	if err != nil {
		return fmt.Errorf("error decoding assets/rain-sound.wav for speaker initialization: %w", err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	return nil
}

// playRainSound plays the continuous rain sound.
func playRainSound(volume float64) {
	f, err := os.Open("assets/rain-sound.wav")
	if err != nil {
		log.Printf("Error opening assets/rain-sound.wav for rain sound: %v\n", err)
		return
	}
	// Do NOT defer f.Close() here, as the streamer needs the file to remain open for looping.

	streamer, _, err := wav.Decode(f)
	if err != nil {
		log.Printf("Error decoding assets/rain-sound.wav for rain sound: %v\n", err)
		f.Close()
		return
	}
	// Do NOT defer streamer.Close() here, as the looped streamer will be managed by the speaker.

	vol := &effects.Volume{
		Streamer: beep.Loop(-1, streamer),
		Base:     2,
		Volume:   math.Log2(volume),
		Silent:   false,
	}
	speaker.Play(vol)
}

// playThunderSound plays a single thunder sound.
func playThunderSound(volume float64) {
	f, err := os.Open("assets/lightning.wav")
	if err != nil {
		log.Printf("Error opening assets/lightning.wav: %v\n", err)
		return
	}
	defer f.Close()

	streamer, _, err := wav.Decode(f)
	if err != nil {
		log.Printf("Error decoding assets/lightning.wav: %v\n", err)
		return
	}
	defer streamer.Close()

	vol := &effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   math.Log2(volume),
		Silent:   false,
	}

	speaker.Play(vol)
}

// setupColors initializes styles for rain and lightning based on user-defined colors.
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

// updateBolt updates the lightning bolt's segments and growth status.
// Returns true if the bolt should continue to exist (i.e., has active segments).
func (bolt *LightningBolt) update() bool {
	currentTime := time.Now()

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
				offset := rand.Intn(5) - 2
				nextX := max(0, min(bolt.maxX-1, currentX+offset))
				nextY := min(bolt.maxY-1, lastY+1)
				newSegments = append(newSegments, LightningSegment{nextY, nextX, currentTime})
				if i == 0 {
					nextPrimaryX = nextX
				}
				currentX = nextX
				addedSegment = true
			}

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
		} else {
			bolt.isGrowing = false
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

// drawBolt draws the lightning bolt on the screen, fading segments based on age.
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

// Helper functions for min/max
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

// simulateRain runs the main simulation loop for rain and lightning effects.
func simulateRain(screen tcell.Screen, muteFlag bool, volume float64, rainColor, lightningColor string) error {
	if !muteFlag {
		go playRainSound(volume)
	}

	setupColors(rainColor, lightningColor)
	rand.Seed(time.Now().UnixNano())
	raindrops := []Raindrop{}
	activeBolts := []*LightningBolt{}
	isThunderstorm := false

	events := make(chan tcell.Event, 10)
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	lastUpdateTime := time.Now()
	for {
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
				case tcell.KeyEscape, tcell.KeyCtrlC:
					return nil
				}
			case *tcell.EventResize:
				screen.Clear()
				raindrops = nil
				activeBolts = nil
			}
		default:
			currentTime := time.Now()
			deltaTime := currentTime.Sub(lastUpdateTime)
			if deltaTime < updateInterval {
				time.Sleep(updateInterval - deltaTime)
			}
			lastUpdateTime = time.Now()

			maxX, maxY := screen.Size()
			var nextBolts []*LightningBolt
			if isThunderstorm && len(activeBolts) < 3 && rand.Float64() < lightningChance {
				go playThunderSound(volume)

				startCol := rand.Intn(maxX/2) + maxX/4
				startRow := rand.Intn(maxY / 5)
				// Corrected targetLength calculation to match Python's random.randint(min, max) behavior
				// Go's rand.Intn(n) gives [0, n-1]. To get [A, B], use rand.Intn(B - A + 1) + A.
				// Here, A = maxY/2, B = maxY - 2.
				targetLength := rand.Intn((maxY-2)-(maxY/2)+1) + (maxY / 2)
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
			nextRaindrops := raindrops[:0]
			for _, drop := range raindrops {
				drop.y += drop.speed
				if int(drop.y) < maxY {
					nextRaindrops = append(nextRaindrops, drop)
				}
			}
			raindrops = nextRaindrops

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

// runApp contains the main logic of the application, returning an error if something goes wrong.
func runApp() error {
	// Command-line flags for customization
	versionFlag := flag.Bool("version", false, "Print version information and exit") // New version flag
	volume := flag.Float64("volume", 1.0, "Volume level for rain and thunder sounds (0.0 to 1.0)")
	muteFlag := flag.Bool("mute", false, "Disable rain and thunder sounds")
	rainColor := flag.String("rain-color", "cyan", "Color for the rain (black, red, green, yellow, blue, magenta, cyan, white)")
	lightningColor := flag.String("lightning-color", "yellow", "Color for the lightning (black, red, green, yellow, blue, magenta, cyan, white)")
	flag.Parse()

	if *versionFlag {
		fmt.Println("GoRain Simulator v1.0.1 - Audio Fixes & Version Flag") // Version string
		return nil                                                          // Exit after printing version
	}

	// Initialize the audio speaker once at the start of the program.
	if err := initAudio(); err != nil {
		return fmt.Errorf("failed to initialize audio: %w", err)
	}
	defer speaker.Close() // Ensure the speaker is closed when runApp exits

	// Validate if the program is running in a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("this program requires a TTY with terminal support")
	}

	// Initialize tcell screen for terminal UI
	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("error initializing screen: %w", err)
	}
	if err := screen.Init(); err != nil {
		return fmt.Errorf("error initializing screen: %w", err)
	}
	defer screen.Fini() // Ensure the screen is finalized when runApp exits

	// Run the rain simulation
	if err := simulateRain(screen, *muteFlag, *volume, *rainColor, *lightningColor); err != nil {
		return fmt.Errorf("error during simulation: %w", err)
	}
	return nil
}

func main() {
	// Set up logging to a file
	logFile, err := os.OpenFile("gorain.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Run the application and handle any errors
	if err := runApp(); err != nil {
		// Print critical errors to stderr so the user sees them
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}

/*package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"
)

// Configuration constants
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

// Color mapping
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

	// Growth
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

			// Add secondary forks
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

	// Check for removal
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

// Helper functions
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

	// Event channel for input
	events := make(chan tcell.Event)
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	lastUpdateTime := time.Now()
	for {
		// Handle input
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
			// Frame rate control
			currentTime := time.Now()
			deltaTime := currentTime.Sub(lastUpdateTime)
			if deltaTime < updateInterval {
				time.Sleep(updateInterval - deltaTime)
			}
			lastUpdateTime = time.Now()

			// Update lightning
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
	// Command-line flags
	rainColor := flag.String("rain-color", "cyan", "Color for the rain (black, red, green, yellow, blue, magenta, cyan, white)")
	lightningColor := flag.String("lightning-color", "yellow", "Color for the lightning (black, red, green, yellow, blue, magenta, cyan, white)")
	flag.Parse()

	// Validate terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		println("Error: This program requires a TTY with terminal support.")
		os.Exit(1)
	}

	// Initialize tcell
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

	// Run simulation
	if err := simulateRain(screen, *rainColor, *lightningColor); err != nil {
		screen.Fini()
		println("Error during simulation:", err)
		os.Exit(1)
	}
}
*/
