.PHONY: all run build clean visualizer random gui

# デフォルトターゲット
all: build

# ビルド
build:
	go build -o bin/musl Main.go

# 実行
run: build
	./bin/musl

# ビジュアライザーの実行
visualizer:
	go run cmd/visualizer/main.go

# GUIビジュアライザーの実行
gui:
	./run_gui.sh

# ランダムデータのビジュアライゼーション
random:
	go run cmd/visualizer/main.go

# クリーン
clean:
	rm -rf bin/
	rm -rf genre_frames/
	rm -rf random_frames/
	rm -f genre_evolution.gif
	rm -f random_evolution.gif 