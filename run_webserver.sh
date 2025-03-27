#!/bin/bash

echo "MuSL Webビジュアライザセットアップを開始します..."

# 必要なパッケージのインストール
echo "必要なGoパッケージをインストールしています..."
go get -u github.com/gorilla/mux
go get -u github.com/gorilla/websocket

echo "ディレクトリを作成しています..."
mkdir -p cmd/webserver
mkdir -p static

# ファイルの存在確認
if [ ! -f "cmd/webserver/handlers.go" ]; then
    echo "エラー: cmd/webserver/handlers.go が見つかりません"
    echo "handlers.go ファイルが正しく作成されているか確認してください"
    exit 1
fi

echo "Webサーバーを起動します: http://localhost:8080"
go run cmd/webserver/*.go 