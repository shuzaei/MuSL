package main

import (
	"fmt"
	"strconv"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"
)

// シミュレーション設定タブを作成
func createSimulationTab(window fyne.Window) *fyne.Container {
	// ステータスとプログレスバーの設定
	statusLabel := widget.NewLabel("準備完了")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()
	
	// グローバル変数にコールバック関数を設定
	mainWindow = window
	statusCallback = func(status string) {
		statusLabel.SetText(status)
	}
	progressCallback = func(progress float64) {
		if progress <= 0 {
			progressBar.Hide()
		} else {
			progressBar.Show()
			progressBar.SetValue(progress)
		}
	}
	
	// シミュレーションパラメータ
	majorProbLabel := widget.NewLabel("メジャーイベント確率:")
	majorProbSlider := widget.NewSlider(0.0, 1.0)
	majorProbSlider.SetValue(0.3) // デフォルト値
	majorProbValue := widget.NewLabel("0.30")
	majorProbSlider.OnChanged = func(v float64) {
		majorProbValue.SetText(fmt.Sprintf("%.2f", v))
	}
	
	initialAgentsLabel := widget.NewLabel("初期エージェント数:")
	initialAgentsEntry := widget.NewEntry()
	initialAgentsEntry.SetText("100")
	
	iterationsLabel := widget.NewLabel("イテレーション数:")
	iterationsEntry := widget.NewEntry()
	iterationsEntry.SetText("100")
	
	// 実行ボタン
	runButton := widget.NewButton("シミュレーション実行", func() {
		majorProb := majorProbSlider.Value
		agents, _ := strconv.Atoi(initialAgentsEntry.Text)
		iters, _ := strconv.Atoi(iterationsEntry.Text)
		
		// シミュレーション実行関数を呼び出す
		go runSimulation(majorProb, agents, iters)
	})
	
	// レイアウト作成
	form := container.New(layout.NewFormLayout(),
		majorProbLabel, container.NewBorder(nil, nil, nil, majorProbValue, majorProbSlider),
		initialAgentsLabel, initialAgentsEntry,
		iterationsLabel, iterationsEntry,
	)
	
	// 説明テキスト
	infoText := widget.NewMultiLineEntry()
	infoText.SetText("MuSLシミュレーションの設定パラメータ：\n\n" +
		"・メジャーイベント確率: 大規模なイベントが発生する確率\n" +
		"・初期エージェント数: シミュレーション開始時のエージェント数\n" +
		"・イテレーション数: シミュレーションの実行回数\n\n" +
		"「シミュレーション実行」ボタンをクリックすると、設定されたパラメータでシミュレーションが実行され、結果が保存されます。")
	infoText.Disable()
	
	// 状態表示領域
	statusContainer := container.NewVBox(
		statusLabel,
		progressBar,
	)
	
	// 全体レイアウト
	content := container.NewBorder(
		form, 
		statusContainer, 
		runButton, nil, 
		container.NewScroll(infoText),
	)
	
	return content
}

// ビジュアライズタブを作成
func createVisualizationTab() *fyne.Container {
	// フレーム生成設定
	frameLabel := widget.NewLabel("フレーム生成:")
	frameButton := widget.NewButton("フレーム生成", func() {
		// フレーム生成関数呼び出し
		go generateFrames()
	})
	
	// GIF作成設定
	gifLabel := widget.NewLabel("GIF作成:")
	gifButton := widget.NewButton("GIF作成", func() {
		// GIF作成関数呼び出し
		go createGIF()
	})
	
	// 一括処理
	allButton := widget.NewButton("フレーム生成＆GIF作成", func() {
		// フレーム生成とGIF作成を順番に実行
		go func() {
			generateFrames()
			createGIF()
		}()
	})
	
	// Webビジュアライザー起動
	webButton := widget.NewButton("Webビジュアライザー起動", func() {
		// サーバー起動関数呼び出し
		go runWebServer()
	})
	
	// 現在の進行状況表示エリア
	progress := widget.NewProgressBar()
	progress.Hide()
	statusLabel := widget.NewLabel("")
	
	// コントロールとステータスのレイアウト
	controls := container.NewVBox(
		container.New(layout.NewFormLayout(),
			frameLabel, frameButton,
			gifLabel, gifButton,
		),
		widget.NewSeparator(),
		allButton,
		widget.NewSeparator(),
		webButton,
		widget.NewSeparator(),
		statusLabel,
		progress,
	)
	
	// 説明テキスト
	infoText := widget.NewMultiLineEntry()
	infoText.SetText(
		"ビジュアライゼーション操作：\n\n" +
		"・フレーム生成: シミュレーション結果からジャンル分布のフレームを生成します\n" +
		"・GIF作成: 生成したフレームからGIFアニメーションを作成します\n" +
		"・一括処理: フレーム生成とGIF作成を連続して実行します\n" +
		"・Webビジュアライザー: インタラクティブなWebビジュアライザーを起動します\n\n" +
		"処理中はステータスが表示されます。",
	)
	infoText.Disable()
	
	return container.NewHSplit(
		controls,
		container.NewScroll(infoText),
	)
}

// ランダム比較タブを作成
func createRandomComparisonTab() *fyne.Container {
	// ランダムデータパラメーター
	initialSongsLabel := widget.NewLabel("初期曲数:")
	initialSongsEntry := widget.NewEntry()
	initialSongsEntry.SetText("50")
	
	framesLabel := widget.NewLabel("フレーム数:")
	framesEntry := widget.NewEntry()
	framesEntry.SetText("100")
	
	// ランダム分布タイプ選択
	distTypeLabel := widget.NewLabel("分布タイプ:")
	distTypeSelect := widget.NewSelect(
		[]string{"完全一様分布", "正規分布", "偏りあり分布"}, 
		func(selected string) {}
	)
	distTypeSelect.SetSelected("完全一様分布")
	
	// 実行ボタン
	runButton := widget.NewButton("ランダムデータGIF生成", func() {
		initialSongs, _ := strconv.Atoi(initialSongsEntry.Text)
		frames, _ := strconv.Atoi(framesEntry.Text)
		distType := distTypeSelect.Selected
		
		// ランダムデータ処理を実行
		go generateRandomVisualization(initialSongs, frames, distType)
	})
	
	// シミュレーションとの比較ボタン
	compareButton := widget.NewButton("シミュレーションと比較", func() {
		// 比較表示関数を呼び出し
		go compareSimulationWithRandom()
	})
	
	// フォームレイアウト
	form := container.New(layout.NewFormLayout(),
		initialSongsLabel, initialSongsEntry,
		framesLabel, framesEntry,
		distTypeLabel, distTypeSelect,
	)
	
	// ボタンレイアウト
	buttons := container.NewVBox(
		runButton,
		widget.NewSeparator(),
		compareButton,
	)
	
	// 説明テキスト
	infoText := widget.NewMultiLineEntry()
	infoText.SetText(
		"ランダムデータとの比較機能：\n\n" +
		"・初期曲数: ランダムデータ生成に使用する初期曲数\n" +
		"・フレーム数: 生成するアニメーションのフレーム数\n" +
		"・分布タイプ: ランダムデータの生成方法\n" +
		"  - 完全一様分布: 完全にランダムな分布\n" +
		"  - 正規分布: 中心付近に集中する分布\n" +
		"  - 偏りあり分布: 年代に応じて偏りがある分布\n\n" +
		"「ランダムデータGIF生成」ボタンを押すと、指定した設定でランダムなジャンル進化のGIFを作成します。\n" +
		"「シミュレーションと比較」ボタンを押すと、実際のシミュレーション結果とランダムデータを並べて表示します。"
	)
	infoText.Disable()	
	
	return container.NewBorder(
		form, 
		buttons, 
		nil, 
		nil, 
		container.NewScroll(infoText),
	)
} 