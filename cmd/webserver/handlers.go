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
	"math"
	"math/rand"
	"math/sort"
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

	// 全曲のIDを取得して一貫した色付けのために使用
	allSongIds := getAllSongIDs(summaries)

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

		// 枠線を描画
		drawBorder(img, 5, color.RGBA{200, 200, 200, 255})

		// 色の凡例を描画
		legendX := width - 150
		legendY := height - 80
		legendWidth := 120
		legendHeight := 20

		// 凡例タイトル
		drawSimpleText(img, "曲の古さ:", legendX, legendY-15, color.RGBA{0, 0, 0, 255})

		// グラデーションバー
		for x := 0; x < legendWidth; x++ {
			normalizedPos := float64(x) / float64(legendWidth-1)
			// 古い曲（青）から新しい曲（赤）へのグラデーション
			r := uint8((0.4 + normalizedPos*0.6) * 255)
			g := uint8((0.4 - normalizedPos*0.25) * 255)
			b := uint8((1.0 - normalizedPos*0.7) * 255)
			a := uint8((0.6 + normalizedPos*0.4) * 255)

			legendColor := color.RGBA{r, g, b, a}

			for y := 0; y < legendHeight; y++ {
				if legendX+x >= 0 && legendX+x < width && legendY+y >= 0 && legendY+y < height {
					img.Set(legendX+x, legendY+y, legendColor)
				}
			}
		}

		// 凡例ラベル
		drawSimpleText(img, "古い", legendX, legendY+legendHeight+15, color.RGBA{0, 0, 200, 255})
		drawSimpleText(img, "新しい", legendX+legendWidth-50, legendY+legendHeight+15, color.RGBA{200, 0, 0, 255})

		// ジャンルポイントをプロット
		if len(summary.AllGenres) > 0 {
			for _, genreInfo := range summary.AllGenres {
				if len(genreInfo.Genre) >= numDimensions {
					// ジャンルの最初の2次元を使用して座標を計算
					x := int(genreInfo.Genre[0]*float64(width-40)) + 20
					y := int(genreInfo.Genre[1]*float64(height-40)) + 20

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

					// 曲の古さに基づいて色を生成
					pointColor := getSongColorByAge(genreInfo.ID, allSongIds)

					// 曲の古さに基づいてサイズを決定（新しい曲ほど大きく）
					pointSize := getSongSizeByAge(genreInfo.ID, allSongIds)

					// 点を描画
					drawPoint(img, x, y, pointSize, pointColor)
				}
			}
		}

		// イテレーション番号を表示
		iterNumStr := fmt.Sprintf("Iteration: %d", i)
		drawSimpleText(img, iterNumStr, 20, 20, color.RGBA{0, 0, 0, 255})

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
					// 中心ほど濃く（グラデーション）
					dist := math.Sqrt(float64(dx*dx+dy*dy)) / float64(size)
					alpha := uint8(float64(c.A) * (1.0 - 0.7*dist))
					if alpha > 0 {
						blended := color.RGBA{c.R, c.G, c.B, alpha}
						img.Set(px, py, blended)
					}
				}
			}
		}
	}
}

// 枠線を描画する関数
func drawBorder(img *image.RGBA, thickness int, c color.RGBA) {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	// 上辺
	for y := 0; y < thickness; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}

	// 下辺
	for y := h - thickness; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}

	// 左辺
	for x := 0; x < thickness; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, c)
		}
	}

	// 右辺
	for x := w - thickness; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, c)
		}
	}
}

// シンプルなテキスト描画関数
func drawSimpleText(img *image.RGBA, text string, x, y int, c color.RGBA) {
	// 非常に簡易的な実装 - 実際のアプリケーションではフォントレンダリングライブラリを使用
	for i, char := range text {
		drawSimpleChar(img, char, x+i*8, y, c)
	}
}

// シンプルな文字描画関数
func drawSimpleChar(img *image.RGBA, char rune, x, y int, c color.RGBA) {
	// 基本的な文字の形状を描画するだけの非常に簡易的な実装
	bounds := img.Bounds()
	if x < 0 || x >= bounds.Max.X || y < 0 || y >= bounds.Max.Y {
		return
	}

	// 文字の形をドットで表現
	switch char {
	case 'I', 'i', 'l', '1':
		for dy := -5; dy <= 2; dy++ {
			img.Set(x, y+dy, c)
		}
	case 'T', 't':
		for dx := -2; dx <= 2; dx++ {
			img.Set(x+dx, y-5, c)
		}
		for dy := -4; dy <= 2; dy++ {
			img.Set(x, y+dy, c)
		}
	case ':':
		img.Set(x, y-3, c)
		img.Set(x, y, c)
	default:
		// その他の文字は単純な点で表現
		img.Set(x, y, c)
	}
}

// 曲の古さに基づいて色を決定する関数
func getSongColorByAge(songID int, allIDs []int) color.RGBA {
	// IDリスト内での位置を探す
	position := 0
	for i, id := range allIDs {
		if id == songID {
			position = i
			break
		}
	}

	// 位置を0〜1の範囲に正規化
	normalizedPos := float64(position) / float64(len(allIDs)-1)
	if math.IsNaN(normalizedPos) {
		normalizedPos = 0.5 // 曲が1つしかない場合のフォールバック
	}

	// 古い曲（青）から新しい曲（赤）へのグラデーション
	r := uint8((0.4 + normalizedPos*0.6) * 255)
	g := uint8((0.4 - normalizedPos*0.25) * 255)
	b := uint8((1.0 - normalizedPos*0.7) * 255)
	a := uint8((0.6 + normalizedPos*0.4) * 255)

	return color.RGBA{r, g, b, a}
}

// 曲の古さに基づいてサイズを決定する関数
func getSongSizeByAge(songID int, allIDs []int) int {
	// IDリスト内での位置を探す
	position := 0
	for i, id := range allIDs {
		if id == songID {
			position = i
			break
		}
	}

	// 位置を0〜1の範囲に正規化
	normalizedPos := float64(position) / float64(len(allIDs)-1)
	if math.IsNaN(normalizedPos) {
		normalizedPos = 0.5 // 曲が1つしかない場合のフォールバック
	}

	// 新しい曲ほど大きく（3〜6ピクセル）
	return 3 + int(normalizedPos*3)
}

// 全曲のIDを取得する関数
func getAllSongIDs(summaries []PublicSummery) []int {
	idMap := make(map[int]bool)

	// 全てのサマリーから曲IDを収集
	for _, summary := range summaries {
		for _, genreInfo := range summary.AllGenres {
			idMap[genreInfo.ID] = true
		}
	}

	// マップをスライスに変換
	ids := make([]int, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}

	// IDを昇順にソート（古い順）
	sort.Ints(ids)

	return ids
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

// ランダムなジャンルフレームを生成する関数
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

	// 分布タイプに基づいて点を生成
	songIDs := make([]int, initialSongs)
	songGenres := make([][]float64, initialSongs)

	// 各曲に一貫したIDと初期ジャンルを割り当て
	for i := 0; i < initialSongs; i++ {
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
	for i := 0; i < numFrames; i++ {
		// 進捗状況の更新
		if numFrames > 0 {
			currentProgress.Progress = 0.3 + 0.6*float64(i)/float64(numFrames)
			currentProgress.Message = fmt.Sprintf("ランダムフレーム生成中... (%d/%d)", i+1, numFrames)
		}

		// 各イテレーションの曲を準備
		activeSongs := make([]int, 0)
		activeGenres := make([][]float64, 0)

		// 活性化する曲を選択
		activationCount := initialSongs / 2
		if i < 10 {
			// 初期フレームは少数の曲から始める
			activationCount = int(float64(initialSongs) * float64(i+1) / 20.0)
		}

		// 少なくとも1曲は表示
		if activationCount < 1 {
			activationCount = 1
		}

		// 曲をシャッフル
		indices := rand.Perm(initialSongs)
		for j := 0; j < activationCount && j < len(indices); j++ {
			idx := indices[j]
			activeSongs = append(activeSongs, songIDs[idx])
			activeGenres = append(activeGenres, songGenres[idx])
		}

		// フレーム画像を生成
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// 背景を白で塗りつぶす
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		// 枠線を描画
		drawBorder(img, 5, color.RGBA{200, 200, 200, 255})

		// 色の凡例を描画
		legendX := width - 150
		legendY := height - 80
		legendWidth := 120
		legendHeight := 20

		// 凡例タイトル
		drawSimpleText(img, "曲の古さ:", legendX, legendY-15, color.RGBA{0, 0, 0, 255})

		// グラデーションバー
		for x := 0; x < legendWidth; x++ {
			normalizedPos := float64(x) / float64(legendWidth-1)
			// 古い曲（青）から新しい曲（赤）へのグラデーション
			r := uint8((0.4 + normalizedPos*0.6) * 255)
			g := uint8((0.4 - normalizedPos*0.25) * 255)
			b := uint8((1.0 - normalizedPos*0.7) * 255)
			a := uint8((0.6 + normalizedPos*0.4) * 255)

			legendColor := color.RGBA{r, g, b, a}

			for y := 0; y < legendHeight; y++ {
				if legendX+x >= 0 && legendX+x < width && legendY+y >= 0 && legendY+y < height {
					img.Set(legendX+x, legendY+y, legendColor)
				}
			}
		}

		// 凡例ラベル
		drawSimpleText(img, "古い", legendX, legendY+legendHeight+15, color.RGBA{0, 0, 200, 255})
		drawSimpleText(img, "新しい", legendX+legendWidth-50, legendY+legendHeight+15, color.RGBA{200, 0, 0, 255})

		// ジャンル点をプロット
		for j, id := range activeSongs {
			if j < len(activeGenres) {
				genre := activeGenres[j]
				if len(genre) >= 2 {
					// 座標計算
					x := int(genre[0]*float64(width-40)) + 20
					y := int(genre[1]*float64(height-40)) + 20

					// 範囲チェック
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

					// 曲の古さに基づいて色を生成
					pointColor := getSongColorByAge(id, songIDs)

					// 曲の古さに基づいてサイズを決定（新しい曲ほど大きく）
					pointSize := getSongSizeByAge(id, songIDs)

					// 点を描画
					drawPoint(img, x, y, pointSize, pointColor)
				}
			}
		}

		// フレーム情報の描画
		iterText := fmt.Sprintf("Iteration: %d (RANDOM - %s)", i, distType)
		drawSimpleText(img, iterText, 20, 20, color.RGBA{0, 0, 0, 255})

		// アクティブな曲数の表示
		songCountText := fmt.Sprintf("Active Songs: %d", len(activeSongs))
		drawSimpleText(img, songCountText, 20, 40, color.RGBA{0, 0, 0, 255})

		// PNG保存
		filename := filepath.Join(frameDir, fmt.Sprintf("frame_%04d.png", i))
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("フレーム保存エラー: %w", err)
		}

		if err := png.Encode(file, img); err != nil {
			file.Close()
			return fmt.Errorf("PNG作成エラー: %w", err)
		}

		file.Close()
	}

	currentProgress.Progress = 0.9
	currentProgress.Message = "GIF作成中..."

	// GIF作成
	if err := createGIFFromFrames("random_frames", "random_evolution.gif", 10); err != nil {
		return fmt.Errorf("GIF作成エラー: %w", err)
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
