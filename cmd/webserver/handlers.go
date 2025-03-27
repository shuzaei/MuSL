package main

import (
	"MuSL/MuSL"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// 進捗状況を保持する構造体
type ProgressStatus struct {
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
	Message  string  `json:"message"`
}

// 現在の進捗状況
var currentProgress ProgressStatus = ProgressStatus{
	Status:   "idle",
	Progress: 0.0,
	Message:  "準備完了",
}

// PublicSummeryはMuSL.PublicSummeryと互換性を持つための構造体
type PublicSummery struct {
	NumAgents    int         `json:"num_population"`
	NumCreators  int         `json:"num_creaters"`
	NumListeners int         `json:"num_listeners"`
	NumSongs     int         `json:"num_song_now"`
	AllGenres    []GenreInfo `json:"all_genres"`
}

// GenreInfoはジャンル情報とIDを持つ構造体
type GenreInfo struct {
	ID    int       `json:"id"`
	Genre []float64 `json:"genre"`
}

// シミュレーション実行ハンドラ
func runSimulationHandler(w http.ResponseWriter, r *http.Request) {
	// リクエストデータを解析
	var requestData struct {
		MajorProb     float64 `json:"majorProb"`
		InitialAgents int     `json:"initialAgents"`
		Iterations    int     `json:"iterations"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// 進捗状況の初期化
	currentProgress = ProgressStatus{
		Status:   "running",
		Progress: 0.0,
		Message:  "シミュレーション開始...",
	}

	// GAパラメータの設定
	gaParams := MuSL.MakeGAParams(0.1, 0.05)

	// デフォルトエージェントパラメータの設定
	defaultAgentParams := MuSL.MakeNewAgent(
		-1,                       // id
		[]bool{true, true, true}, // role
		100.0,                    // energy
		100.0,                    // default_energy
		0.0,                      // elimination_threshold
		0.5,                      // reproduction_probability

		// creator
		0.5,                   // innovation_rate
		make([]*MuSL.Song, 0), // memory
		0.5,                   // creation_probability
		1.0,                   // creation_cost

		// listener
		0.5,                    // novelty_preference
		make([]*MuSL.Song, 0),  // memory
		make([]*MuSL.Song, 0),  // incoming_songs
		make([]*MuSL.Event, 0), // song_events
		0.5,                    // listening_probability
		1.0,                    // evaluation_cost

		// organizer
		MuSL.Const64(requestData.MajorProb), // major_probability
		make([]*MuSL.Event, 0),              // created_events
		0.5,                                 // event_probability
		0.5,                                 // organization_cost
		1.0,                                 // organization_reward

		// メジャーイベントパラメータ
		0.5, // major_listener_ratio
		0.5, // major_creator_ratio
		0.1, // major_song_ratio
		0.5, // major_winner_ratio
		0.5, // major_reward_ratio
		0.1, // major_recommendation_ratio

		// マイナーイベントパラメータ
		0.1, // minor_listener_ratio
		0.1, // minor_creator_ratio
		0.5, // minor_song_ratio
		0.5, // minor_reward_ratio
		0.1, // minor_recommendation_ratio
	)

	// シミュレーション作成
	sim := MuSL.MakeNewSimulation(requestData.InitialAgents, requestData.Iterations, gaParams, defaultAgentParams)

	// 進捗報告を行いながらシミュレーション実行
	go func() {
		// シミュレーション実行
		sim.Run()

		currentProgress.Progress = 0.8
		currentProgress.Message = "シミュレーション完了、結果を保存中..."

		// サマリー取得
		summaries := sim.GetSummery()

		// 結果をJSONで保存
		outputJSON, err := json.Marshal(summaries)
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "JSONエンコードエラー: " + err.Error()
			return
		}

		err = ioutil.WriteFile("output.json", outputJSON, 0644)
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "ファイル保存エラー: " + err.Error()
			return
		}

		currentProgress.Status = "completed"
		currentProgress.Progress = 1.0
		currentProgress.Message = "シミュレーション完了！output.jsonに保存されました"
	}()

	// すぐに応答を返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "シミュレーションを開始しました",
	})
}

// 進捗状況取得ハンドラ
func getProgressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentProgress)
}

// フレーム生成ハンドラ
func generateFramesHandler(w http.ResponseWriter, r *http.Request) {
	// 進捗状況の初期化
	currentProgress = ProgressStatus{
		Status:   "running",
		Progress: 0.1,
		Message:  "フレーム生成開始...",
	}

	go func() {
		// output.jsonからデータ読み込み
		data, err := ioutil.ReadFile("output.json")
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "output.jsonが見つかりません。先にシミュレーションを実行してください。"
			return
		}

		var summaries []PublicSummery
		err = json.Unmarshal(data, &summaries)
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "JSONデコードエラー: " + err.Error()
			return
		}

		currentProgress.Progress = 0.3
		currentProgress.Message = fmt.Sprintf("%d イテレーションのデータを読み込みました", len(summaries))

		// フレームディレクトリ作成
		frameDir := "genre_frames"
		err = os.MkdirAll(frameDir, 0755)
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "ディレクトリ作成エラー: " + err.Error()
			return
		}

		// フレーム生成
		currentProgress.Progress = 0.4
		currentProgress.Message = "フレーム生成中..."

		// フレーム生成処理
		generateGenreFrames(summaries, frameDir)

		currentProgress.Status = "completed"
		currentProgress.Progress = 1.0
		currentProgress.Message = "フレーム生成完了！"
	}()

	// すぐに応答を返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "フレーム生成を開始しました",
	})
}

// フレーム生成関数
func generateGenreFrames(summaries []PublicSummery, frameDir string) {
	const (
		width         = 800
		height        = 600
		numDimensions = 2 // 2次元の場合
	)

	// フレーム生成
	for i, summary := range summaries {
		// 進捗状況の更新
		if len(summaries) > 0 {
			currentProgress.Progress = 0.4 + 0.5*float64(i)/float64(len(summaries))
			currentProgress.Message = fmt.Sprintf("フレーム生成中... (%d/%d)", i+1, len(summaries))
		}

		// 新しい画像を作成
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// 背景を白で塗りつぶす
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		// ジャンルポイントをプロット
		if len(summary.AllGenres) > 0 {
			for _, genreInfo := range summary.AllGenres {
				if len(genreInfo.Genre) >= numDimensions {
					// ジャンルの最初の2次元を使用して座標を計算
					x := int(genreInfo.Genre[0]*float64(width-20)) + 10
					y := int(genreInfo.Genre[1]*float64(height-20)) + 10

					// 描画範囲内に収まるように調整
					if x < 0 {
						x = 0
					} else if x >= width {
						x = width - 1
					}

					if y < 0 {
						y = 0
					} else if y >= height {
						y = height - 1
					}

					// IDをハッシュ化して一貫した色を生成
					r := uint8((genreInfo.ID * 123) % 255)
					g := uint8((genreInfo.ID * 45) % 255)
					b := uint8((genreInfo.ID * 67) % 255)

					// 点を描画 (3x3ピクセルの円形)
					pointColor := color.RGBA{r, g, b, 255}
					drawPoint(img, x, y, 3, pointColor)
				}
			}
		}

		// イテレーション番号をレンダリング
		// 実際の実装では、テキスト描画ライブラリを使用するとよい

		// ファイル名の生成とPNG形式で保存
		filename := filepath.Join(frameDir, fmt.Sprintf("frame_%04d.png", i))
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("フレーム保存エラー: %s\n", err)
			continue
		}

		// PNGとして保存
		if err := png.Encode(file, img); err != nil {
			file.Close()
			fmt.Printf("PNG作成エラー: %s\n", err)
			continue
		}

		file.Close()
	}
}

// 点を描画する関数
func drawPoint(img *image.RGBA, x, y, size int, c color.RGBA) {
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			// 円形の点を描画
			if dx*dx+dy*dy <= size*size {
				px, py := x+dx, y+dy
				if px >= 0 && px < img.Bounds().Max.X && py >= 0 && py < img.Bounds().Max.Y {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// GIF作成ハンドラ
func createGIFHandler(w http.ResponseWriter, r *http.Request) {
	// 進捗状況の初期化
	currentProgress = ProgressStatus{
		Status:   "running",
		Progress: 0.1,
		Message:  "GIF作成開始...",
	}

	go func() {
		// フレームディレクトリの確認
		frameDir := "genre_frames"
		if _, err := os.Stat(frameDir); os.IsNotExist(err) {
			currentProgress.Status = "error"
			currentProgress.Message = "genre_framesディレクトリが見つかりません。先にフレームを生成してください。"
			return
		}

		// PNG画像の取得
		pattern := filepath.Join(frameDir, "*.png")
		files, err := filepath.Glob(pattern)
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "ファイル検索エラー: " + err.Error()
			return
		}

		if len(files) == 0 {
			currentProgress.Status = "error"
			currentProgress.Message = "PNG画像が見つかりません。先にフレームを生成してください。"
			return
		}

		currentProgress.Progress = 0.2
		currentProgress.Message = fmt.Sprintf("%d個のPNGファイルを処理します...", len(files))

		// GIF作成処理
		outputFile := "genre_evolution.gif"
		err = createGIFFromFrames(frameDir, outputFile, 20)
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "GIF作成エラー: " + err.Error()
			return
		}

		currentProgress.Status = "completed"
		currentProgress.Progress = 1.0
		currentProgress.Message = "GIF作成完了！"
	}()

	// すぐに応答を返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "GIF作成を開始しました",
	})
}

// GIF作成関数
func createGIFFromFrames(frameDir, outputFile string, delay int) error {
	// フレームファイルの取得
	pattern := filepath.Join(frameDir, "*.png")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// ファイルを数値順にソート
	sort.Slice(files, func(i, j int) bool {
		// ファイル名から数値部分を抽出
		baseI := filepath.Base(files[i])
		baseJ := filepath.Base(files[j])

		// "frame_"とextensionを除去
		numStrI := baseI[6 : len(baseI)-4]
		numStrJ := baseJ[6 : len(baseJ)-4]

		// 数値に変換
		numI, _ := strconv.Atoi(numStrI)
		numJ, _ := strconv.Atoi(numStrJ)

		return numI < numJ
	})

	if len(files) == 0 {
		return fmt.Errorf("フレームが見つかりません")
	}

	// 最初の画像からGIFのサイズを取得
	firstFile, err := os.Open(files[0])
	if err != nil {
		return err
	}
	firstImg, err := png.Decode(firstFile)
	firstFile.Close()
	if err != nil {
		return err
	}

	bounds := firstImg.Bounds()

	// GIFアニメーションの作成
	anim := gif.GIF{}

	// 標準パレットを作成（256色）
	stdPalette := make(color.Palette, 256)
	stdPalette[0] = color.RGBA{255, 255, 255, 255} // 白（背景色）

	// パレットに基本色を追加
	for i := 0; i < 216; i++ {
		r := uint8((i / 36) * 51)
		g := uint8(((i / 6) % 6) * 51)
		b := uint8((i % 6) * 51)
		stdPalette[i+1] = color.RGBA{r, g, b, 255}
	}

	// グレースケールを追加
	for i := 0; i < 24; i++ {
		gray := uint8(i * 10)
		stdPalette[i+217] = color.RGBA{gray, gray, gray, 255}
	}

	// 残りは黒で埋める
	for i := 241; i < 256; i++ {
		stdPalette[i] = color.RGBA{0, 0, 0, 255}
	}

	// 各フレームを読み込んでGIFに追加
	for i, file := range files {
		// 進捗状況の更新
		if len(files) > 0 {
			currentProgress.Progress = 0.3 + 0.6*float64(i)/float64(len(files))
			currentProgress.Message = fmt.Sprintf("GIF作成中... (%d/%d)", i+1, len(files))
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			return err
		}

		// パレット画像に変換
		palettedImg := image.NewPaletted(bounds, stdPalette)
		draw.Draw(palettedImg, bounds, img, bounds.Min, draw.Over)

		// フレームを追加
		anim.Image = append(anim.Image, palettedImg)
		anim.Delay = append(anim.Delay, delay/10) // 100分の1秒単位
	}

	// GIFファイルに保存
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return gif.EncodeAll(outFile, &anim)
}

// ランダムビジュアライゼーションハンドラ
func randomVisualizationHandler(w http.ResponseWriter, r *http.Request) {
	// リクエストデータを解析
	var requestData struct {
		InitialSongs int    `json:"initialSongs"`
		Frames       int    `json:"frames"`
		DistType     string `json:"distType"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// 進捗状況の初期化
	currentProgress = ProgressStatus{
		Status:   "running",
		Progress: 0.1,
		Message:  "ランダムデータ生成開始...",
	}

	go func() {
		// ランダムデータフレーム生成
		currentProgress.Progress = 0.3
		currentProgress.Message = "ランダムデータフレーム生成中..."

		err := generateRandomFrames(requestData.InitialSongs, requestData.Frames, requestData.DistType)
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "フレーム生成エラー: " + err.Error()
			return
		}

		// GIF作成
		currentProgress.Progress = 0.7
		currentProgress.Message = "ランダムデータGIF作成中..."

		err = createGIFFromFrames("random_frames", "random_evolution.gif", 20)
		if err != nil {
			currentProgress.Status = "error"
			currentProgress.Message = "GIF作成エラー: " + err.Error()
			return
		}

		currentProgress.Status = "completed"
		currentProgress.Progress = 1.0
		currentProgress.Message = "ランダムデータGIF作成完了！"
	}()

	// すぐに応答を返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "ランダムデータGIF生成を開始しました",
	})
}

// ランダムデータ生成関数
func generateRandomFrames(initialSongs, numFrames int, distType string) error {
	// ランダムデータ用のディレクトリ作成
	frameDir := "random_frames"
	if err := os.MkdirAll(frameDir, 0755); err != nil {
		return err
	}

	const (
		width  = 800
		height = 600
	)

	// 全フレームで使用する曲のIDとジャンル情報
	maxSongs := initialSongs * 3 // 最大曲数（徐々に増加させる）
	songIDs := make([]int, maxSongs)
	songGenres := make([][]float64, maxSongs)

	// 全曲に一貫したIDと固定のジャンル位置を割り当て
	for i := 0; i < maxSongs; i++ {
		songIDs[i] = i + 1

		var x, y float64
		switch distType {
		case "uniform":
			// 一様分布
			x = rand.Float64()
			y = rand.Float64()
		case "normal":
			// 正規分布（中心に集中）
			x = rand.NormFloat64()*0.15 + 0.5
			y = rand.NormFloat64()*0.15 + 0.5
			// 0-1の範囲に収める
			x = clamp(x, 0, 1)
			y = clamp(y, 0, 1)
		case "biased":
			// 偏りのある分布（数か所に集中）
			center := rand.Intn(3)
			switch center {
			case 0:
				x = rand.NormFloat64()*0.1 + 0.2
				y = rand.NormFloat64()*0.1 + 0.2
			case 1:
				x = rand.NormFloat64()*0.1 + 0.5
				y = rand.NormFloat64()*0.1 + 0.5
			case 2:
				x = rand.NormFloat64()*0.1 + 0.8
				y = rand.NormFloat64()*0.1 + 0.8
			}
			// 0-1の範囲に収める
			x = clamp(x, 0, 1)
			y = clamp(y, 0, 1)
		default:
			// デフォルトは一様分布
			x = rand.Float64()
			y = rand.Float64()
		}

		songGenres[i] = []float64{x, y}
	}

	// フレーム生成
	// 各フレームで表示する曲数を徐々に増やす
	for frame := 0; frame < numFrames; frame++ {
		// 進捗状況の更新
		if numFrames > 0 {
			currentProgress.Progress = 0.3 + 0.4*float64(frame)/float64(numFrames)
			currentProgress.Message = fmt.Sprintf("ランダムデータフレーム生成中... (%d/%d)", frame+1, numFrames)
		}

		// 新しい画像を作成
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// 背景を白で塗りつぶす
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		// このフレームで表示する曲数を計算（徐々に増加）
		currentSongCount := initialSongs
		if frame > 0 {
			// 最初のフレームからinitialSongs曲を表示し、その後徐々に増加
			additionalSongs := (maxSongs - initialSongs) * frame / (numFrames - 1)
			currentSongCount = initialSongs + additionalSongs
			if currentSongCount > maxSongs {
				currentSongCount = maxSongs
			}
		}

		// 各曲のジャンルポイントをプロット（固定位置）
		for i := 0; i < currentSongCount; i++ {
			genre := songGenres[i]
			id := songIDs[i]

			x := int(genre[0]*float64(width-20)) + 10
			y := int(genre[1]*float64(height-20)) + 10

			// IDをハッシュ化して一貫した色を生成
			r := uint8((id * 123) % 255)
			g := uint8((id * 45) % 255)
			b := uint8((id * 67) % 255)

			// 点を描画
			pointColor := color.RGBA{r, g, b, 255}
			drawPoint(img, x, y, 3, pointColor)
		}

		// フレーム上部にフレーム番号と曲数を表示
		// 実際のテキスト描画には別のライブラリが必要なため、ここではコメントのみ
		// drawText(img, 10, 20, fmt.Sprintf("Frame: %d, Songs: %d", frame, currentSongCount))

		// ファイル名の生成とPNG形式で保存
		filename := filepath.Join(frameDir, fmt.Sprintf("frame_%04d.png", frame))
		file, err := os.Create(filename)
		if err != nil {
			return err
		}

		// PNGとして保存
		if err := png.Encode(file, img); err != nil {
			file.Close()
			return err
		}

		file.Close()
	}

	return nil
}

// 値を範囲内に収める関数
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// 時系列グラフデータ取得ハンドラ
func getTimeSeriesDataHandler(w http.ResponseWriter, r *http.Request) {
	// output.jsonからデータ読み込み
	data, err := ioutil.ReadFile("output.json")
	if err != nil {
		http.Error(w, "output.jsonが見つかりません。先にシミュレーションを実行してください。", http.StatusNotFound)
		return
	}

	var summaries []PublicSummery
	err = json.Unmarshal(data, &summaries)
	if err != nil {
		http.Error(w, "JSONデコードエラー: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 時系列データの準備
	timeSeriesData := generateTimeSeriesData(summaries)

	// JSONとして返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeSeriesData)
}

// 時系列データ生成関数
func generateTimeSeriesData(summaries []PublicSummery) map[string]interface{} {
	iterations := len(summaries)

	// 各イテレーションでのデータポイント
	timeSteps := make([]int, iterations)
	numAgents := make([]int, iterations)
	numCreators := make([]int, iterations)
	numListeners := make([]int, iterations)
	numSongs := make([]int, iterations)

	// ジャンル特性の分析用データ
	genreDiversity := make([]float64, iterations)
	genreClusters := make([]int, iterations)

	for i, summary := range summaries {
		timeSteps[i] = i
		numAgents[i] = summary.NumAgents
		numCreators[i] = summary.NumCreators
		numListeners[i] = summary.NumListeners
		numSongs[i] = summary.NumSongs

		// ジャンルの多様性を計算（簡易版：ユニークなジャンルの数を数える）
		genreMap := make(map[string]bool)
		for _, genreInfo := range summary.AllGenres {
			genreKey := fmt.Sprintf("%.2f-%.2f", genreInfo.Genre[0], genreInfo.Genre[1])
			genreMap[genreKey] = true
		}
		genreDiversity[i] = float64(len(genreMap)) / float64(max(1, summary.NumSongs))

		// クラスター数の近似（単純化のため、実際のクラスタリングアルゴリズムは実装していない）
		// 実装する場合はk-meansなどを使用するとよい
		genreClusters[i] = estimateClusters(summary.AllGenres)
	}

	// 結果を構造化
	return map[string]interface{}{
		"timeSteps": timeSteps,
		"population": map[string]interface{}{
			"total":     numAgents,
			"creators":  numCreators,
			"listeners": numListeners,
		},
		"songs": numSongs,
		"genres": map[string]interface{}{
			"diversity": genreDiversity,
			"clusters":  genreClusters,
		},
	}
}

// クラスター数を推定する簡易関数（実際には正確なクラスタリングを行うべき）
func estimateClusters(genres []GenreInfo) int {
	if len(genres) <= 1 {
		return len(genres)
	}

	// 非常に単純化した実装：
	// 2次元空間を3x3のグリッドに分割し、それぞれのセルに曲が存在するかをカウント
	grid := make([][]bool, 3)
	for i := range grid {
		grid[i] = make([]bool, 3)
	}

	for _, genreInfo := range genres {
		if len(genreInfo.Genre) >= 2 {
			x := int(genreInfo.Genre[0] * 3)
			y := int(genreInfo.Genre[1] * 3)

			// 境界チェック
			if x >= 3 {
				x = 2
			}
			if y >= 3 {
				y = 2
			}

			grid[x][y] = true
		}
	}

	// グリッドのうち、曲が存在するセルの数をカウント
	clusterCount := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if grid[i][j] {
				clusterCount++
			}
		}
	}

	return clusterCount
}

// 最大値を取得するヘルパー関数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
