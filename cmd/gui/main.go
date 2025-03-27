package main

import (
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func init() {
	// 乱数のシード値を設定
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// FyneアプリケーションとウィンドウのPDO作成
	a := app.New()

	// テーマの設定（オプション）
	// a.Settings().SetTheme(theme.LightTheme())

	// メインウィンドウの作成
	w := a.NewWindow("MuSL ビジュアライザー")
	w.Resize(fyne.NewSize(800, 600))

	// タブコンテナの作成
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("シミュレーション", theme.DocumentCreateIcon(), createSimulationTab(w)),
		container.NewTabItemWithIcon("視覚化", theme.ViewFullScreenIcon(), createVisualizationTab()),
		container.NewTabItemWithIcon("ランダム比較", theme.ViewRestoreIcon(), createRandomComparisonTab()),
	)

	// タブの位置を上部に設定
	tabs.SetTabLocation(container.TabLocationTop)

	// ウィンドウにタブを設定
	w.SetContent(tabs)

	// 終了時のクリーンアップ
	w.SetCloseIntercept(func() {
		// 必要に応じてクリーンアップ処理を実装

		// 通常の終了処理を呼び出す
		w.Close()
	})

	// アプリケーションの開始
	fmt.Println("MuSL GUIアプリケーションを起動しています...")
	w.ShowAndRun()
}
