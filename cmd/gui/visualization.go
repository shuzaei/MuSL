package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// PublicSummeryはMuSL.PublicSummeryと互換性を持つための構造体
type PublicSummery struct {
	NumAgents    int
	NumCreators  int
	NumListeners int
	NumSongs     int
	AllGenres    []GenreInfo
}

// GenreInfoはジャンル情報とIDを持つ構造体
type GenreInfo struct {
	ID    int
	Genre []float64
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

					// 点を描画 (5x5ピクセルの四角形)
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

	// 各フレームを読み込んでGIFに追加
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			return err
		}

		// パレット化
		palettedImg := image.NewPaletted(bounds, nil)
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
	for frame := 0; frame < numFrames; frame++ {
		// 新しい画像を作成
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// 背景を白で塗りつぶす
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		// 各曲のジャンルポイントをプロット
		for i := 0; i < initialSongs; i++ {
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

			// ランダムな動き（小さな揺らぎ）を適用
			if frame < numFrames-1 {
				// 次のフレームに向けてジャンルを少し変更
				dx := (rand.Float64() - 0.5) * 0.02
				dy := (rand.Float64() - 0.5) * 0.02

				songGenres[i][0] = clamp(songGenres[i][0]+dx, 0, 1)
				songGenres[i][1] = clamp(songGenres[i][1]+dy, 0, 1)
			}
		}

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
