package main

import (
	"MuSL/MuSL"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// グローバルステート
var (
	statusCallback   func(string)
	progressCallback func(float64)
	mainWindow       fyne.Window
)

// シミュレーションを実行する関数
func runSimulation(majorProb float64, initialAgents, iterations int) {
	updateStatus("シミュレーション開始...")
	updateProgress(0.0)

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
		MuSL.Const64(majorProb), // major_probability
		make([]*MuSL.Event, 0),  // created_events
		0.5,                     // event_probability
		0.5,                     // organization_cost
		1.0,                     // organization_reward

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
	sim := MuSL.MakeNewSimulation(initialAgents, iterations, gaParams, defaultAgentParams)

	// シミュレーション実行
	sim.Run()

	// サマリー取得
	summaries := sim.GetSummery()

	// 結果をJSONで保存
	outputJSON, err := json.Marshal(summaries)
	if err != nil {
		showError("JSONエンコードエラー", err.Error())
		return
	}

	err = ioutil.WriteFile("output.json", outputJSON, 0644)
	if err != nil {
		showError("ファイル保存エラー", err.Error())
		return
	}

	updateStatus("シミュレーション完了！output.jsonに保存されました")
	updateProgress(1.0)

	// 完了ダイアログ
	showInfo("シミュレーション完了", fmt.Sprintf("%d イテレーションのシミュレーションが完了しました。\noutput.jsonに結果が保存されました。", iterations))
}

// GenreInfoの変換関数
func convertGenreInfos(genres []MuSL.GenreInfo) []GenreInfo {
	result := make([]GenreInfo, len(genres))
	for i, g := range genres {
		result[i] = GenreInfo{
			ID:    g.ID,
			Genre: g.Genre,
		}
	}
	return result
}

// フレーム生成関数
func generateFrames() {
	updateStatus("フレーム生成開始...")
	updateProgress(0.1)

	// output.jsonからデータ読み込み
	data, err := ioutil.ReadFile("output.json")
	if err != nil {
		showError("ファイル読み込みエラー", "output.jsonが見つかりません。先にシミュレーションを実行してください。")
		return
	}

	var summaries []PublicSummery
	err = json.Unmarshal(data, &summaries)
	if err != nil {
		showError("JSONデコードエラー", err.Error())
		return
	}

	updateStatus(fmt.Sprintf("%d イテレーションのデータを読み込みました", len(summaries)))
	updateProgress(0.2)

	// フレームディレクトリ作成
	frameDir := "genre_frames"
	err = os.MkdirAll(frameDir, 0755)
	if err != nil {
		showError("ディレクトリ作成エラー", err.Error())
		return
	}

	// フレーム生成
	updateStatus("フレーム生成中...")
	generateGenreFrames(summaries, frameDir)

	updateProgress(1.0)
	updateStatus("フレーム生成完了！")
	showInfo("処理完了", "genre_framesディレクトリにフレームが生成されました。")
}

// GIF作成関数
func createGIF() {
	updateStatus("GIF作成開始...")
	updateProgress(0.1)

	// フレームディレクトリの確認
	frameDir := "genre_frames"
	if _, err := os.Stat(frameDir); os.IsNotExist(err) {
		showError("ディレクトリエラー", "genre_framesディレクトリが見つかりません。先にフレームを生成してください。")
		return
	}

	// PNG画像の取得
	pattern := filepath.Join(frameDir, "*.png")
	files, err := filepath.Glob(pattern)
	if err != nil {
		showError("ファイル検索エラー", err.Error())
		return
	}

	if len(files) == 0 {
		showError("ファイルなし", "PNG画像が見つかりません。先にフレームを生成してください。")
		return
	}

	updateStatus(fmt.Sprintf("%d個のPNGファイルを処理します...", len(files)))
	updateProgress(0.2)

	// GIF作成処理
	outputFile := "genre_evolution.gif"
	err = createGIFFromFrames(frameDir, outputFile, 20)
	if err != nil {
		showError("GIF作成エラー", err.Error())
		return
	}

	updateProgress(1.0)
	updateStatus("GIF作成完了！")
	showInfo("処理完了", fmt.Sprintf("GIFアニメーションが%sに作成されました", outputFile))
}

// Webサーバー起動関数
func runWebServer() {
	updateStatus("Webサーバー起動中...")

	// ポート設定
	port := "8080"

	// ファイルサーバーハンドラー作成
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// サーバー情報表示
	url := fmt.Sprintf("http://localhost:%s/visualizer.html", port)
	updateStatus(fmt.Sprintf("サーバー起動: %s", url))

	// ブラウザでURLを開く
	openBrowser(url)

	// サーバー起動
	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			showError("サーバーエラー", err.Error())
		}
	}()
}

// ランダムデータビジュアライゼーション生成
func generateRandomVisualization(initialSongs, frames int, distType string) {
	updateStatus("ランダムデータ生成開始...")
	updateProgress(0.1)

	// ランダムデータのマッピング
	distTypeMap := map[string]string{
		"完全一様分布": "uniform",
		"正規分布":   "normal",
		"偏りあり分布": "biased",
	}

	mappedDistType := distTypeMap[distType]
	if mappedDistType == "" {
		mappedDistType = "uniform" // デフォルト
	}

	// ランダムデータフレーム生成
	updateStatus("ランダムデータフレーム生成中...")
	updateProgress(0.3)

	err := generateRandomFrames(initialSongs, frames, mappedDistType)
	if err != nil {
		showError("フレーム生成エラー", err.Error())
		return
	}

	// GIF作成
	updateStatus("ランダムデータGIF作成中...")
	updateProgress(0.7)

	err = createGIFFromFrames("random_frames", "random_evolution.gif", 20)
	if err != nil {
		showError("GIF作成エラー", err.Error())
		return
	}

	updateProgress(1.0)
	updateStatus("ランダムデータGIF作成完了！")
	showInfo("処理完了", "random_evolution.gifが生成されました。")
}

// シミュレーションとランダムデータを比較表示
func compareSimulationWithRandom() {
	updateStatus("比較表示準備中...")

	// 両方のGIFが存在するか確認
	if _, err := os.Stat("genre_evolution.gif"); os.IsNotExist(err) {
		showError("ファイルなし", "genre_evolution.gifが見つかりません。シミュレーションとフレーム生成を先に実行してください。")
		return
	}

	if _, err := os.Stat("random_evolution.gif"); os.IsNotExist(err) {
		showError("ファイルなし", "random_evolution.gifが見つかりません。ランダムデータGIF生成を先に実行してください。")
		return
	}

	// 比較表示用のHTMLファイル生成
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>MuSL Simulation vs Random Data</title>
		<style>
			body { font-family: Arial, sans-serif; text-align: center; }
			.container { display: flex; justify-content: center; }
			.gif-container { margin: 20px; }
		</style>
	</head>
	<body>
		<h1>MuSLシミュレーション vs ランダムデータ比較</h1>
		<div class="container">
			<div class="gif-container">
				<h2>シミュレーションデータ</h2>
				<img src="genre_evolution.gif" alt="Simulation Data">
			</div>
			<div class="gif-container">
				<h2>ランダムデータ</h2>
				<img src="random_evolution.gif" alt="Random Data">
			</div>
		</div>
	</body>
	</html>
	`

	err := ioutil.WriteFile("comparison.html", []byte(htmlContent), 0644)
	if err != nil {
		showError("ファイル作成エラー", err.Error())
		return
	}

	// ブラウザで表示
	updateStatus("比較表示を開きます...")
	openBrowser("comparison.html")
}

// ブラウザでURLを開く関数
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("このプラットフォームではブラウザを自動で開けません: %s", runtime.GOOS)
	}

	if err != nil {
		showError("ブラウザ起動エラー", err.Error())
	}
}

// ステータス更新関数
func updateStatus(status string) {
	if statusCallback != nil {
		statusCallback(status)
	}
}

// 進捗更新関数
func updateProgress(progress float64) {
	if progressCallback != nil {
		progressCallback(progress)
	}
}

// エラーダイアログ表示
func showError(title, message string) {
	if mainWindow != nil {
		dialog.ShowError(fmt.Errorf(message), mainWindow)
	} else {
		log.Printf("エラー: %s - %s", title, message)
	}
}

// 情報ダイアログ表示
func showInfo(title, message string) {
	if mainWindow != nil {
		dialog.ShowInformation(title, message, mainWindow)
	} else {
		log.Printf("情報: %s - %s", title, message)
	}
}
