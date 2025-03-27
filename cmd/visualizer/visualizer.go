package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenreInfo represents a song genre with its ID
type GenreInfo struct {
	ID    int       `json:"id"`
	Genre []float64 `json:"genre"`
}

// PublicSummery represents the data structure from the simulation
type PublicSummery struct {
	NumPopulation        int         `json:"num_population"`
	NumCreaters          int         `json:"num_creaters"`
	NumListeners         int         `json:"num_listeners"`
	NumOrganizers        int         `json:"num_organizers"`
	NumSongAll           int         `json:"num_song_all"`
	NumSongThis          int         `json:"num_song_this"`
	NumSongNow           int         `json:"num_song_now"`
	NumEvaluationAll     int         `json:"num_evaluation_all"`
	NumEvaluationThis    int         `json:"num_evaluation_this"`
	NumEventAll          int         `json:"num_event_all"`
	NumEventThis         int         `json:"num_event_this"`
	NumDeaths            int         `json:"num_deaths"`        // 死者数
	NumReproductions     int         `json:"num_reproductions"` // 再生産数
	AvgInnovation        float64     `json:"avg_innovation"`
	AvgNoveltyPreference float64     `json:"avg_novelty_preference"`
	SumEvaluation        float64     `json:"sum_evaluation"`
	AvgEvaluation        float64     `json:"avg_evaluation"`
	TotalEnergy          float64     `json:"total_energy"`
	EnergyCreators       float64     `json:"energy_creators"`
	EnergyListeners      float64     `json:"energy_listeners"`
	EnergyOrganizers     float64     `json:"energy_organizers"`
	AllGenres            []GenreInfo `json:"all_genres"`
}

func main() {
	fmt.Println("\nMuSL Visualization Menu")
	fmt.Println("======================")

	var choice string
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nOptions:")
		fmt.Println("1: Web visualizer (opens a web browser with interactive charts)")
		fmt.Println("2: Generate genre evolution frames")
		fmt.Println("3: Create GIF from frames")
		fmt.Println("4: Generate frames AND create GIF (options 2+3)")
		fmt.Println("q: Quit")
		fmt.Print("\nEnter your choice: ")

		choice, _ = reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Println("\nStarting web visualizer...")
			go func() {
				err := runWebServer()
				if err != nil {
					fmt.Printf("Error running web server: %v\n", err)
				}
			}()
			fmt.Println("Web server started at http://localhost:8080/visualizer.html")
			fmt.Println("Press Ctrl+C to stop the server when done.")
			return

		case "2":
			fmt.Println("\nGenerating genre evolution frames...")
			generateFrames()

		case "3":
			fmt.Println("\nCreating GIF from frames...")
			createGIF()

		case "4":
			fmt.Println("\nGenerating frames and creating GIF...")
			generateFrames()
			createGIF()

		case "q", "Q", "quit", "exit":
			fmt.Println("\nExiting...")
			return

		default:
			fmt.Println("\nInvalid choice. Please try again.")
		}
	}
}

func runWebServer() error {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a file server handler
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// Print server info
	log.Printf("Starting server on :%s", port)
	log.Printf("Access visualizer at http://localhost:%s/visualizer.html", port)

	// Start server
	return http.ListenAndServe(":"+port, nil)
}

func generateFrames() {
	// Load data from output.json
	data, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatalf("Error reading output.json: %v", err)
	}

	var summaries []PublicSummery
	err = json.Unmarshal(data, &summaries)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	fmt.Printf("Loaded %d iterations of data\n", len(summaries))

	// Create output directory
	frameDir := "genre_frames"
	generateGenreFrames(summaries, frameDir)
	fmt.Println("Genre evolution frames created in the genre_frames directory")
}

func createGIF() {
	// Directory containing PNG frames
	frameDir := "genre_frames"

	// Output GIF file
	outputFile := "genre_evolution.gif"

	// Check if frames directory exists
	if _, err := os.Stat(frameDir); os.IsNotExist(err) {
		log.Fatalf("Frames directory '%s' does not exist. Generate frames first.", frameDir)
	}

	// Get all PNG files
	pattern := filepath.Join(frameDir, "*.png")
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatalf("Error finding PNG files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No PNG files found in '%s'", frameDir)
	}

	// Sort files by name to ensure correct order
	sort.Strings(files)

	fmt.Printf("Found %d PNG files\n", len(files))

	// Create GIF
	outGif := &gif.GIF{}

	// Process each file
	for i, file := range files {
		// Skip non-PNG files
		if !strings.HasSuffix(strings.ToLower(file), ".png") {
			continue
		}

		// Open the PNG file
		f, err := os.Open(file)
		if err != nil {
			log.Printf("Error opening file %s: %v", file, err)
			continue
		}

		// Decode PNG
		img, err := png.Decode(f)
		if err != nil {
			f.Close()
			log.Printf("Error decoding PNG %s: %v", file, err)
			continue
		}
		f.Close()

		// Convert to paletted image
		bounds := img.Bounds()
		palettedImg := image.NewPaletted(bounds, palette.Plan9)

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				palettedImg.Set(x, y, img.At(x, y))
			}
		}

		// Add to GIF
		outGif.Image = append(outGif.Image, palettedImg)
		outGif.Delay = append(outGif.Delay, 20) // 1/5 second delay
		outGif.Disposal = append(outGif.Disposal, gif.DisposalNone)

		// Print progress
		if i%10 == 0 || i == len(files)-1 {
			fmt.Printf("Processed %d/%d files\n", i+1, len(files))
		}
	}

	// Save the GIF
	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}

	err = gif.EncodeAll(f, outGif)
	if err != nil {
		f.Close()
		log.Fatalf("Error encoding GIF: %v", err)
	}

	f.Close()
	fmt.Printf("Created GIF animation: %s\n", outputFile)
}

// getAllSongIDs collects all song IDs across all iterations
func getAllSongIDs(summaries []PublicSummery) []int {
	idMap := make(map[int]bool)

	for _, summary := range summaries {
		for _, genreInfo := range summary.AllGenres {
			idMap[genreInfo.ID] = true
		}
	}

	// Convert map to slice
	ids := make([]int, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}

	// Sort IDs for consistency
	sort.Ints(ids)

	return ids
}

// getSongColor returns a color for a song based on its ID
// 新しい曲ほど明るくて目立つ色になるように変更
func getSongColor(songID int, allIDs []int, totalIDs int) color.RGBA {
	// Find the relative position of this ID in the list
	position := 0
	for i, id := range allIDs {
		if id == songID {
			position = i
			break
		}
	}

	// Calculate normalized position (0.0-1.0)
	normalizedPos := float64(position) / float64(totalIDs-1)
	if math.IsNaN(normalizedPos) {
		normalizedPos = 0.5 // Fallback if only one song
	}

	// IDが大きいほど新しい曲 (新しい曲ほど鮮やかな色に)
	// 逆にして新しい曲が1.0に近くなるようにする
	brightness := 0.3 + (1.0-normalizedPos)*0.7 // 新曲が明るく(100%)、古い曲が暗く(30%)

	// 色相はレインボーカラーを維持
	hue := normalizedPos * 360.0 // 0-360 degrees

	// Convert HSV to RGB
	var r, g, b float64

	// Calculate RGB from HSV
	h := hue / 60.0
	i := math.Floor(h)
	f := h - i

	v := brightness // 明るさ
	s := 0.8        // 彩度を固定

	p := v * (1.0 - s)
	q := v * (1.0 - s*f)
	t := v * (1.0 - s*(1.0-f))

	switch int(i) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	case 5:
		r, g, b = v, p, q
	}

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// generateGenreFrames creates PNG frames showing the evolution of genres
func generateGenreFrames(summaries []PublicSummery, outputDir string) {
	// Create output directory if it doesn't exist
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Get all song IDs across all iterations to assign consistent colors
	allSongIDs := getAllSongIDs(summaries)
	fmt.Printf("Found %d unique songs in the simulation\n", len(allSongIDs))

	width := 800
	height := 800
	margin := 50

	// 死者数と再生産数の累積値を追跡
	cumulativeDeaths := 0
	cumulativeReproductions := 0

	// Process each iteration/frame
	for i, summary := range summaries {
		// Skip frames with no genre data
		if len(summary.AllGenres) == 0 {
			continue
		}

		// 累積値を更新
		cumulativeDeaths += summary.NumDeaths
		cumulativeReproductions += summary.NumReproductions

		// Create a new image
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// Fill background with white
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, color.White)
			}
		}

		// Draw axes (black lines)
		// X-axis
		for x := margin; x < width-margin; x++ {
			img.Set(x, height-margin, color.Black)
		}

		// Y-axis
		for y := margin; y < height-margin; y++ {
			img.Set(margin, y, color.Black)
		}

		// Draw grid lines (light gray)
		gridColor := color.RGBA{220, 220, 220, 255}
		gridSpacing := float64(width-2*margin) / 10.0

		// Horizontal grid lines
		for i := 1; i < 10; i++ {
			y := int(float64(margin) + float64(i)*gridSpacing)
			for x := margin; x < width-margin; x++ {
				img.Set(x, y, gridColor)
			}
		}

		// Vertical grid lines
		for i := 1; i < 10; i++ {
			x := int(float64(margin) + float64(i)*gridSpacing)
			for y := margin; y < height-margin; y++ {
				img.Set(x, y, gridColor)
			}
		}

		// Draw data points
		scale := float64(width - 2*margin)

		// Count songs for opacity
		totalSongs := len(summary.AllGenres)
		if totalSongs == 0 {
			continue
		}

		// Draw each genre point
		for _, genreInfo := range summary.AllGenres {
			// Skip invalid genres
			if len(genreInfo.Genre) < 2 {
				continue
			}

			// Map genre dimensions to coordinates
			xf := float64(margin) + genreInfo.Genre[0]*scale
			yf := float64(height-margin) - genreInfo.Genre[1]*scale

			x := int(xf)
			y := int(yf)

			// Get color based on song ID (consistent across all frames)
			pointColor := getSongColor(genreInfo.ID, allSongIDs, len(allSongIDs))

			// Draw a simple point (5x5 square)
			pointSize := 2
			for py := y - pointSize; py <= y+pointSize; py++ {
				for px := x - pointSize; px <= x+pointSize; px++ {
					if px >= 0 && px < width && py >= 0 && py < height {
						img.Set(px, py, pointColor)
					}
				}
			}
		}

		// Draw frame information at the top
		frameInfoY := 15

		// Draw iteration number
		iterText := fmt.Sprintf("Iteration: %d", i)
		drawText(img, iterText, margin, frameInfoY, color.RGBA{0, 0, 0, 255})

		// Draw statistics
		statsX := width - margin - 200
		statsY := margin + 15
		statsLineHeight := 15

		drawText(img, fmt.Sprintf("Songs: %d", summary.NumSongAll), statsX, statsY, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("Population: %d", summary.NumPopulation), statsX, statsY+statsLineHeight, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("Deaths: %d", cumulativeDeaths), statsX, statsY+2*statsLineHeight, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("Reproductions: %d", cumulativeReproductions), statsX, statsY+3*statsLineHeight, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("Avg Innovation: %.2f", summary.AvgInnovation), statsX, statsY+4*statsLineHeight, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("Avg Novelty Pref: %.2f", summary.AvgNoveltyPreference), statsX, statsY+5*statsLineHeight, color.RGBA{0, 0, 0, 255})

		// Save the image
		outputPath := fmt.Sprintf("%s/frame_%04d.png", outputDir, i)
		f, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}

		// Output frame as PNG
		err = png.Encode(f, img)
		if err != nil {
			f.Close()
			log.Fatalf("Error encoding PNG: %v", err)
		}
		f.Close()

		// Print progress
		if i%10 == 0 {
			fmt.Printf("Processed frame %d/%d\n", i, len(summaries))
		}
	}
}

// Simple text drawing function (very basic, just for labels)
func drawText(img *image.RGBA, text string, x, y int, textColor color.RGBA) {
	// This is a very basic implementation - in a real application you'd use a font rendering library
	for i, char := range text {
		// For each character, draw a simple representation
		drawChar(img, char, x+i*8, y, textColor)
	}
}

// Draw a single character (very simple implementation)
func drawChar(img *image.RGBA, char rune, x, y int, textColor color.RGBA) {
	// Simple lookup table for basic characters
	// This is just a very basic visualization
	switch char {
	case 'I':
		for i := 0; i < 7; i++ {
			img.Set(x+2, y+i-5, textColor)
		}
	case 't':
		for i := 0; i < 5; i++ {
			img.Set(x+2, y+i-5, textColor)
		}
		img.Set(x+1, y-4, textColor)
		img.Set(x+3, y-4, textColor)
	default:
		// For all other characters, just draw a dot
		img.Set(x+2, y-2, textColor)
	}
}
