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
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
		fmt.Println("5: Generate random data frames and GIF (for comparison)")
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

		case "5":
			fmt.Println("\nGenerating random data frames and GIF...")
			generateRandomFrames()
			createRandomGIF()

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

	// 古さに基づいた色分け - より明確にする
	// IDの順序は古い順（IDが小さいほど古い）
	// 最も古い曲: 青系色 (0.6, 0.6, 1.0)
	// 中間の曲: 紫・緑系色
	// 最も新しい曲: 赤系色 (1.0, 0.3, 0.3)

	// より明確な色付けのための調整
	r := 0.4 + normalizedPos*0.6 // 0.4〜1.0（古い曲はやや赤みが少なく、新しい曲ほど赤み強く）
	g := 0.4 - normalizedPos*0.2 // 0.4〜0.2（緑成分は新しくなるほど減少）
	b := 1.0 - normalizedPos*0.8 // 1.0〜0.2（青成分は古い曲ほど強く、新しい曲は弱く）

	// 透明度も年代に応じて調整
	opacity := 0.5 + normalizedPos*0.5 // 50%〜100%（古い曲は半透明で、新しい曲ほど不透明に）

	// 最終的なRGB値を計算
	return color.RGBA{
		R: uint8(r * 255 * opacity),
		G: uint8(g * 255 * opacity),
		B: uint8(b * 255 * opacity),
		A: 255, // 完全不透明
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

	// Process each iteration/frame
	for i, summary := range summaries {
		// Skip frames with no genre data
		if len(summary.AllGenres) == 0 {
			continue
		}

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
		drawText(img, fmt.Sprintf("Avg Innovation: %.2f", summary.AvgInnovation), statsX, statsY+2*statsLineHeight, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("Avg Novelty Pref: %.2f", summary.AvgNoveltyPreference), statsX, statsY+3*statsLineHeight, color.RGBA{0, 0, 0, 255})

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

// generateRandomSummaries creates random data summaries that mimic the real data structure
func generateRandomSummaries(count int, startingSongCount int) []PublicSummery {
	// 現在時刻を使った再現性のないシード値を設定
	rand.Seed(time.Now().UnixNano())
	summaries := make([]PublicSummery, count)

	// 最大でいくつの曲を作るか決める（大幅に増加）
	maxSongs := startingSongCount * 20

	// 全ての曲のIDとジャンルを先に生成
	allSongs := make([]GenreInfo, maxSongs)

	// 各曲に固有のIDとジャンルを割り当て
	for i := 0; i < maxSongs; i++ {
		songID := i + 1

		// 完全な一様ランダム分布
		x := rand.Float64() // 0.0〜1.0の一様乱数
		y := rand.Float64() // 0.0〜1.0の一様乱数

		// ジャンル情報とIDを保存
		allSongs[i] = GenreInfo{
			ID:    songID,
			Genre: []float64{x, y},
		}
	}

	// アクティブな曲のリスト
	activeSongs := make(map[int]GenreInfo)

	// 最初のフレームで表示する曲数（より多くの曲を表示）
	initialActiveSongs := startingSongCount * 2
	if initialActiveSongs < 50 {
		initialActiveSongs = 50
	}

	// 各イテレーションでのジャンルデータを生成
	for i := 0; i < count; i++ {
		// シミュレーション用のサマリー構造体を初期化
		summary := PublicSummery{
			NumPopulation:        100 + rand.Intn(50),
			NumCreaters:          30 + rand.Intn(20),
			NumListeners:         50 + rand.Intn(20),
			NumOrganizers:        10 + rand.Intn(10),
			NumSongAll:           (i+1)*40 + rand.Intn(20), // より多くの曲を累積
			NumSongThis:          10 + rand.Intn(15),       // このターンの新しい曲数も増加
			NumSongNow:           0,                        // 後で計算
			AvgInnovation:        0.3 + rand.Float64()*0.4,
			AvgNoveltyPreference: 0.4 + rand.Float64()*0.3,
			AllGenres:            make([]GenreInfo, 0),
		}

		// 1. アクティブな曲の一部を削除（完全ランダムに）
		if i > 0 && len(activeSongs) > 0 {
			// アクティブな曲の5〜15%をランダムに削除
			removalRate := 0.05 + rand.Float64()*0.1
			removeCount := int(float64(len(activeSongs)) * removalRate)

			// IDのリストを取得
			songIDs := make([]int, 0, len(activeSongs))
			for id := range activeSongs {
				songIDs = append(songIDs, id)
			}

			// 完全にランダムに削除（IDによるバイアスなし）
			rand.Shuffle(len(songIDs), func(i, j int) {
				songIDs[i], songIDs[j] = songIDs[j], songIDs[i]
			})

			// ランダムに選んだ曲を削除
			for j := 0; j < removeCount && j < len(songIDs); j++ {
				delete(activeSongs, songIDs[j])
			}
		}

		// 2. 新しい曲を追加（より多くの曲を追加）
		var addCount int
		if i == 0 {
			addCount = initialActiveSongs
		} else {
			// 各ターンで10〜30の新しい曲を追加（より多くの曲を追加）
			addCount = 10 + rand.Intn(21)

			// 後半のターンでも減少させない（一定の曲数追加を維持）
		}

		// 新しい曲を追加するインデックスを計算
		availableIndices := make([]int, 0, maxSongs)
		for j := 0; j < maxSongs; j++ {
			// 現在アクティブでない曲のインデックスを収集
			if _, exists := activeSongs[allSongs[j].ID]; !exists {
				availableIndices = append(availableIndices, j)
			}
		}

		// 利用可能な曲があれば、ランダムに選択して追加
		if len(availableIndices) > 0 {
			// 追加する数が利用可能な曲数を超えないように調整
			if addCount > len(availableIndices) {
				addCount = len(availableIndices)
			}

			// インデックスをシャッフル
			rand.Shuffle(len(availableIndices), func(i, j int) {
				availableIndices[i], availableIndices[j] = availableIndices[j], availableIndices[i]
			})

			// 新しい曲をランダムに追加
			for j := 0; j < addCount; j++ {
				idx := availableIndices[j]
				activeSongs[allSongs[idx].ID] = allSongs[idx]
			}
		}

		// アクティブな曲を集約
		for _, song := range activeSongs {
			summary.AllGenres = append(summary.AllGenres, song)
		}

		// アクティブな曲数を更新
		summary.NumSongNow = len(summary.AllGenres)

		summaries[i] = summary
	}

	return summaries
}

// generateRandomFrames creates PNG frames with random genre data
func generateRandomFrames() {
	frameCount := 100
	initialSongCount := 50 // 最初のフレームでアクティブになる可能性がある曲数

	// Generate random data
	fmt.Println("Generating random genre data...")
	summaries := generateRandomSummaries(frameCount, initialSongCount)

	// Create output directory
	frameDir := "random_frames"
	err := os.MkdirAll(frameDir, 0755)
	if err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Get all song IDs for consistent coloring
	allSongIDs := getAllSongIDs(summaries)
	fmt.Printf("Generated %d unique random songs\n", len(allSongIDs))

	width := 800
	height := 800
	margin := 50

	// Process each iteration/frame
	for i, summary := range summaries {
		// Create a new image
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// Fill background with white
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, color.White)
			}
		}

		// Draw axes and grid (same as normal frames)
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
		for j := 1; j < 10; j++ {
			y := int(float64(margin) + float64(j)*gridSpacing)
			for x := margin; x < width-margin; x++ {
				img.Set(x, y, gridColor)
			}
		}

		// Vertical grid lines
		for j := 1; j < 10; j++ {
			x := int(float64(margin) + float64(j)*gridSpacing)
			for y := margin; y < height-margin; y++ {
				img.Set(x, y, gridColor)
			}
		}

		// Draw data points
		scale := float64(width - 2*margin)

		// Draw each genre point
		for _, genreInfo := range summary.AllGenres {
			// Map genre dimensions to coordinates
			xf := float64(margin) + genreInfo.Genre[0]*scale
			yf := float64(height-margin) - genreInfo.Genre[1]*scale

			x := int(xf)
			y := int(yf)

			// Get color based on song ID - 一貫した色付け
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

		// Draw frame information
		frameInfoY := 15

		// Draw iteration number
		iterText := fmt.Sprintf("Iteration: %d (RANDOM DATA)", i)
		drawText(img, iterText, margin, frameInfoY, color.RGBA{0, 0, 0, 255})

		// Draw statistics
		statsX := width - margin - 200
		statsY := margin + 15
		statsLineHeight := 15

		drawText(img, fmt.Sprintf("Songs Total: %d", summary.NumSongAll), statsX, statsY, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("Active Songs: %d", summary.NumSongNow), statsX, statsY+statsLineHeight, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("New Songs: %d", summary.NumSongThis), statsX, statsY+2*statsLineHeight, color.RGBA{0, 0, 0, 255})
		drawText(img, "RANDOM SIMULATION", statsX, statsY+4*statsLineHeight, color.RGBA{255, 0, 0, 255})

		// Save the image
		outputPath := fmt.Sprintf("%s/frame_%04d.png", frameDir, i)
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
			fmt.Printf("Processed random frame %d/%d\n", i, len(summaries))
		}
	}

	fmt.Println("Random genre evolution frames created in the random_frames directory")
}

// createRandomGIF creates a GIF from the random frames
func createRandomGIF() {
	// Directory containing PNG frames
	frameDir := "random_frames"

	// Output GIF file
	outputFile := "random_evolution.gif"

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

	fmt.Printf("Found %d random PNG files\n", len(files))

	// Create GIF
	outGif := &gif.GIF{}

	// Process each file
	for i, file := range files {
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
	fmt.Printf("Created random GIF animation: %s\n", outputFile)
}
