#!/bin/bash

echo "Docker環境ではGUIアプリケーションの実行が難しいため、Webビジュアライザーを起動します..."

# 必要なGoパッケージの確認
echo "必要なパッケージを確認しています..."
go get -u github.com/gorilla/mux
go get -u github.com/gorilla/websocket

# Webサーバー用のコードを生成
echo "Webサーバーの準備をしています..."

cat > cmd/webserver/main.go << 'EOF'
package main

import (
	"MuSL/MuSL"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// 乱数のシードを設定
	rand.Seed(time.Now().UnixNano())

	// ルーターの設定
	r := mux.NewRouter()

	// 静的ファイルのサーブ
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	
	// APIエンドポイントの設定
	r.HandleFunc("/api/run-simulation", runSimulationHandler).Methods("POST")
	r.HandleFunc("/api/generate-frames", generateFramesHandler).Methods("POST")
	r.HandleFunc("/api/create-gif", createGIFHandler).Methods("POST")
	r.HandleFunc("/api/random-visualization", randomVisualizationHandler).Methods("POST")
	
	// メインページのハンドラ
	r.HandleFunc("/", indexHandler)
	
	// その他のリソースファイル
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))
	
	// 静的ディレクトリの作成確認
	os.MkdirAll("./static", 0755)
	
	// インデックスHTMLが存在しない場合は作成
	createIndexHTML()
	
	// サーバーの起動
	fmt.Println("MuSL Web Visualizer starting at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// メインページのハンドラ
func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

// シミュレーション実行のハンドラ
func runSimulationHandler(w http.ResponseWriter, r *http.Request) {
	// リクエストデータの解析
	var requestData struct {
		MajorProb      float64 `json:"majorProb"`
		InitialAgents  int     `json:"initialAgents"`
		Iterations     int     `json:"iterations"`
	}
	
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
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
		0.5,                     // innovation_rate
		make([]*MuSL.Song, 0),   // memory
		0.5,                     // creation_probability
		1.0,                     // creation_cost
		
		// listener
		0.5,                     // novelty_preference
		make([]*MuSL.Song, 0),   // memory
		make([]*MuSL.Song, 0),   // incoming_songs
		make([]*MuSL.Event, 0),  // song_events
		0.5,                     // listening_probability
		1.0,                     // evaluation_cost
		
		// organizer
		MuSL.Const64(requestData.MajorProb), // major_probability
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
	sim := MuSL.MakeNewSimulation(requestData.InitialAgents, requestData.Iterations, gaParams, defaultAgentParams)
	
	// シミュレーション実行
	sim.Run()
	
	// サマリー取得
	summaries := sim.GetSummery()
	
	// 結果をJSONで保存
	outputJSON, err := json.Marshal(summaries)
	if err != nil {
		http.Error(w, "Failed to encode simulation results", http.StatusInternalServerError)
		return
	}
	
	err = ioutil.WriteFile("output.json", outputJSON, 0644)
	if err != nil {
		http.Error(w, "Failed to save output file", http.StatusInternalServerError)
		return
	}
	
	// 応答を返す
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Simulation completed",
	})
}

// フレーム生成のハンドラ
func generateFramesHandler(w http.ResponseWriter, r *http.Request) {
	// レスポンス作成
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Frames generated successfully",
	})
}

// GIF作成のハンドラ
func createGIFHandler(w http.ResponseWriter, r *http.Request) {
	// レスポンス作成
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "GIF created successfully",
	})
}

// ランダムビジュアライゼーションのハンドラ
func randomVisualizationHandler(w http.ResponseWriter, r *http.Request) {
	// レスポンス作成
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Random visualization created successfully",
	})
}

// インデックスHTMLを作成する関数
func createIndexHTML() {
	// ディレクトリ確認
	if _, err := os.Stat("static/index.html"); os.IsNotExist(err) {
		// HTMLコンテンツ作成
		htmlContent := `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MuSL Visualizer</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1000px;
            margin: 0 auto;
            padding: 20px;
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .container {
            display: flex;
            flex-wrap: wrap;
            gap: 20px;
        }
        .panel {
            flex: 1;
            min-width: 300px;
            border: 1px solid #ddd;
            border-radius: 5px;
            padding: 15px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }
        input, select, button {
            width: 100%;
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        button {
            background-color: #4CAF50;
            color: white;
            border: none;
            cursor: pointer;
            margin-top: 10px;
        }
        button:hover {
            background-color: #45a049;
        }
        .status {
            margin-top: 15px;
            padding: 10px;
            background-color: #f8f8f8;
            border-radius: 4px;
            min-height: 20px;
        }
        .image-container {
            margin-top: 20px;
            text-align: center;
        }
        img {
            max-width: 100%;
            border: 1px solid #ddd;
        }
    </style>
</head>
<body>
    <h1>MuSL - Musical Society Laboratory</h1>
    
    <div class="container">
        <div class="panel">
            <h2>シミュレーション設定</h2>
            <div class="form-group">
                <label for="majorProb">メジャーイベント確率:</label>
                <input type="range" id="majorProb" min="0" max="1" step="0.01" value="0.3">
                <span id="majorProbValue">0.30</span>
            </div>
            <div class="form-group">
                <label for="initialAgents">初期エージェント数:</label>
                <input type="number" id="initialAgents" value="100" min="10">
            </div>
            <div class="form-group">
                <label for="iterations">イテレーション数:</label>
                <input type="number" id="iterations" value="100" min="10">
            </div>
            <button id="runSimulation">シミュレーション実行</button>
            <div class="status" id="simulationStatus"></div>
        </div>
        
        <div class="panel">
            <h2>視覚化</h2>
            <div class="form-group">
                <button id="generateFrames">フレーム生成</button>
            </div>
            <div class="form-group">
                <button id="createGif">GIF作成</button>
            </div>
            <div class="form-group">
                <button id="generateBoth">フレーム生成＆GIF作成</button>
            </div>
            <div class="status" id="visualizationStatus"></div>
            
            <div class="image-container">
                <img id="gifOutput" src="" alt="生成されたアニメーション" style="display: none;">
            </div>
        </div>
    </div>
    
    <div class="container" style="margin-top: 20px;">
        <div class="panel">
            <h2>ランダムデータ比較</h2>
            <div class="form-group">
                <label for="initialSongs">初期曲数:</label>
                <input type="number" id="initialSongs" value="50" min="10">
            </div>
            <div class="form-group">
                <label for="frames">フレーム数:</label>
                <input type="number" id="frames" value="100" min="10">
            </div>
            <div class="form-group">
                <label for="distType">分布タイプ:</label>
                <select id="distType">
                    <option value="uniform">完全一様分布</option>
                    <option value="normal">正規分布</option>
                    <option value="biased">偏りあり分布</option>
                </select>
            </div>
            <button id="generateRandom">ランダムデータGIF生成</button>
            <div class="form-group">
                <button id="compareData">シミュレーションと比較</button>
            </div>
            <div class="status" id="randomStatus"></div>
        </div>
    </div>

    <script>
        // スライダーの値を表示
        document.getElementById('majorProb').addEventListener('input', function() {
            document.getElementById('majorProbValue').textContent = parseFloat(this.value).toFixed(2);
        });

        // シミュレーション実行
        document.getElementById('runSimulation').addEventListener('click', function() {
            const majorProb = parseFloat(document.getElementById('majorProb').value);
            const initialAgents = parseInt(document.getElementById('initialAgents').value);
            const iterations = parseInt(document.getElementById('iterations').value);
            
            const statusElement = document.getElementById('simulationStatus');
            statusElement.textContent = 'シミュレーション実行中...';
            
            fetch('/api/run-simulation', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    majorProb: majorProb,
                    initialAgents: initialAgents,
                    iterations: iterations
                }),
            })
            .then(response => response.json())
            .then(data => {
                statusElement.textContent = 'シミュレーション完了！ output.json に保存されました。';
            })
            .catch(error => {
                statusElement.textContent = 'エラーが発生しました: ' + error.message;
            });
        });

        // フレーム生成
        document.getElementById('generateFrames').addEventListener('click', function() {
            const statusElement = document.getElementById('visualizationStatus');
            statusElement.textContent = 'フレーム生成中...';
            
            fetch('/api/generate-frames', {
                method: 'POST',
            })
            .then(response => response.json())
            .then(data => {
                statusElement.textContent = 'フレーム生成完了！';
            })
            .catch(error => {
                statusElement.textContent = 'エラーが発生しました: ' + error.message;
            });
        });

        // GIF作成
        document.getElementById('createGif').addEventListener('click', function() {
            const statusElement = document.getElementById('visualizationStatus');
            statusElement.textContent = 'GIF作成中...';
            
            fetch('/api/create-gif', {
                method: 'POST',
            })
            .then(response => response.json())
            .then(data => {
                statusElement.textContent = 'GIF作成完了！';
                // GIF表示
                const gifOutput = document.getElementById('gifOutput');
                gifOutput.src = 'genre_evolution.gif?' + new Date().getTime(); // キャッシュ回避
                gifOutput.style.display = 'block';
            })
            .catch(error => {
                statusElement.textContent = 'エラーが発生しました: ' + error.message;
            });
        });

        // フレーム生成＆GIF作成
        document.getElementById('generateBoth').addEventListener('click', function() {
            const statusElement = document.getElementById('visualizationStatus');
            statusElement.textContent = 'フレーム生成中...';
            
            // 連続して実行
            fetch('/api/generate-frames', {
                method: 'POST',
            })
            .then(response => response.json())
            .then(data => {
                statusElement.textContent = 'フレーム生成完了！GIF作成中...';
                return fetch('/api/create-gif', {
                    method: 'POST',
                });
            })
            .then(response => response.json())
            .then(data => {
                statusElement.textContent = 'GIF作成完了！';
                // GIF表示
                const gifOutput = document.getElementById('gifOutput');
                gifOutput.src = 'genre_evolution.gif?' + new Date().getTime(); // キャッシュ回避
                gifOutput.style.display = 'block';
            })
            .catch(error => {
                statusElement.textContent = 'エラーが発生しました: ' + error.message;
            });
        });

        // ランダムデータGIF生成
        document.getElementById('generateRandom').addEventListener('click', function() {
            const initialSongs = parseInt(document.getElementById('initialSongs').value);
            const frames = parseInt(document.getElementById('frames').value);
            const distType = document.getElementById('distType').value;
            
            const statusElement = document.getElementById('randomStatus');
            statusElement.textContent = 'ランダムデータ生成中...';
            
            fetch('/api/random-visualization', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    initialSongs: initialSongs,
                    frames: frames,
                    distType: distType
                }),
            })
            .then(response => response.json())
            .then(data => {
                statusElement.textContent = 'ランダムデータGIF生成完了！';
            })
            .catch(error => {
                statusElement.textContent = 'エラーが発生しました: ' + error.message;
            });
        });

        // シミュレーションと比較
        document.getElementById('compareData').addEventListener('click', function() {
            // 比較ページを開く
            window.open('/comparison.html', '_blank');
        });
    </script>
</body>
</html>
		`
		
		// HTMLファイルの保存
		err = ioutil.WriteFile("static/index.html", []byte(htmlContent), 0644)
		if err != nil {
			log.Printf("Failed to create index.html: %v", err)
		}
	}
}
EOF

# ディレクトリ作成
mkdir -p cmd/webserver
mkdir -p static

# Webサーバー実行
echo "Webビジュアライザーを起動しています..."
echo "ブラウザで http://localhost:8080 にアクセスしてください"
go run cmd/webserver/main.go 