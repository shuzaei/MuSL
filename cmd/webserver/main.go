package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// 乱数の初期化 (Go 1.20以降はSeedは不要ですが、互換性のため)
	// rand.Seed(time.Now().UnixNano())
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// ルーターの作成
	r := mux.NewRouter()

	// APIルート
	r.HandleFunc("/api/run-simulation", runSimulationHandler).Methods("POST")
	r.HandleFunc("/api/get-progress", getProgressHandler).Methods("GET")
	r.HandleFunc("/api/generate-frames", generateFramesHandler).Methods("POST")
	r.HandleFunc("/api/create-gif", createGIFHandler).Methods("POST")
	r.HandleFunc("/api/random-visualization", randomVisualizationHandler).Methods("POST")
	r.HandleFunc("/api/time-series-data", getTimeSeriesDataHandler).Methods("GET")

	// 静的ファイルサーバー
	staticDir := "./static"
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		log.Fatalf("静的ディレクトリの作成に失敗しました: %v", err)
	}

	// GIFファイルを静的ファイルとして提供
	r.HandleFunc("/genre_evolution.gif", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "genre_evolution.gif")
	})

	r.HandleFunc("/random_evolution.gif", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "random_evolution.gif")
	})

	// HTMLインデックスファイルの作成
	createIndexHTML(staticDir)

	// 静的ファイルの提供
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))

	// サーバー起動
	port := "8080"
	fmt.Printf("Webサーバーを起動中... http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// インデックスHTMLファイルの作成
func createIndexHTML(staticDir string) {
	htmlContent := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MuSL Visualizer</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
    <style>
        body {
            font-family: 'Arial', sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
            color: #333;
        }
        h1, h2 {
            color: #2c3e50;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .tabs {
            display: flex;
            border-bottom: 1px solid #ddd;
            margin-bottom: 20px;
        }
        .tab {
            padding: 10px 20px;
            cursor: pointer;
            border: 1px solid transparent;
            border-bottom: none;
            border-radius: 5px 5px 0 0;
            background-color: #f8f9fa;
            margin-right: 5px;
        }
        .tab.active {
            background-color: white;
            border-color: #ddd;
            border-bottom-color: white;
            font-weight: bold;
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }
        input, select {
            width: 100%;
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
        }
        button {
            background-color: #4CAF50;
            color: white;
            border: none;
            padding: 10px 15px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            margin-right: 10px;
        }
        button:hover {
            background-color: #45a049;
        }
        .info-text {
            background-color: #f8f9fa;
            border-left: 4px solid #4CAF50;
            padding: 10px;
            margin: 15px 0;
        }
        .visualization {
            display: flex;
            flex-direction: column;
            align-items: center;
            margin-top: 20px;
        }
        .gif-display {
            max-width: 100%;
            margin-top: 15px;
            border: 1px solid #ddd;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .progress-container {
            width: 100%;
            background-color: #ddd;
            border-radius: 4px;
            margin: 20px 0;
        }
        .progress-bar {
            height: 20px;
            background-color: #4CAF50;
            border-radius: 4px;
            width: 0%;
            transition: width 0.3s;
        }
        .progress-text {
            margin-top: 5px;
            font-size: 14px;
            color: #666;
        }
        .button-group {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            margin-bottom: 15px;
        }
        .comparison-view {
            display: flex;
            flex-wrap: wrap;
            gap: 20px;
            margin-top: 20px;
        }
        .gif-container {
            flex: 1;
            min-width: 300px;
            display: flex;
            flex-direction: column;
            align-items: center;
        }
        .gif-container h3 {
            margin-bottom: 10px;
        }
        .chart-container {
            margin-bottom: 30px;
            max-width: 100%;
            height: 300px;
        }
        canvas {
            max-width: 100%;
            height: 100%;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>MuSL Visualizer</h1>
        
        <div class="tabs">
            <div class="tab active" data-tab="simulation">シミュレーション</div>
            <div class="tab" data-tab="visualization">可視化</div>
            <div class="tab" data-tab="random-comparison">ランダム比較</div>
            <div class="tab" data-tab="graphs">グラフ分析</div>
        </div>
        
        <div class="tab-content active" id="simulation">
            <h2>シミュレーション設定</h2>
            <div class="form-group">
                <label for="majorProb">メジャーイベント確率:</label>
                <input type="range" id="majorProb" min="0" max="1" step="0.05" value="0.3">
                <span id="majorProbValue">0.3</span>
            </div>
            
            <div class="form-group">
                <label for="initialAgents">初期エージェント数:</label>
                <input type="number" id="initialAgents" value="100" min="10" max="1000">
            </div>
            
            <div class="form-group">
                <label for="iterations">イテレーション数:</label>
                <input type="number" id="iterations" value="100" min="10" max="1000">
            </div>
            
            <button id="runSimulation">シミュレーション実行</button>
            
            <div class="progress-container">
                <div class="progress-bar" id="simulationProgress"></div>
            </div>
            <div class="progress-text" id="simulationStatus">準備完了</div>
            
            <div class="info-text">
                <p><strong>パラメータ説明:</strong></p>
                <p><strong>メジャーイベント確率</strong> - オーガナイザーがメジャーイベントを開催する確率。高いほど業界の集中度が高まります。</p>
                <p><strong>初期エージェント数</strong> - シミュレーション開始時のエージェント（クリエイター、リスナー、オーガナイザー）の数。</p>
                <p><strong>イテレーション数</strong> - シミュレーションの実行回数。多いほど時間の経過とともにシステムがどのように進化するかを詳細に観察できます。</p>
            </div>
        </div>
        
        <div class="tab-content" id="visualization">
            <h2>ジャンル進化の可視化</h2>
            
            <div class="button-group">
                <button id="generateFrames">フレーム生成</button>
                <button id="createGIF">GIF作成</button>
                <button id="runVisualization">一括処理</button>
            </div>
            
            <div class="progress-container">
                <div class="progress-bar" id="visualizationProgress"></div>
            </div>
            <div class="progress-text" id="visualizationStatus">準備完了</div>
            
            <div class="visualization">
                <h3>ジャンル進化アニメーション</h3>
                <img id="genreGIF" class="gif-display" src="" alt="ジャンル進化アニメーション" style="display: none;">
            </div>
            
            <div class="info-text">
                <p><strong>操作説明:</strong></p>
                <p><strong>フレーム生成</strong> - シミュレーション結果からジャンル空間のスナップショットを作成します。</p>
                <p><strong>GIF作成</strong> - 生成されたフレームからアニメーションGIFを作成します。</p>
                <p><strong>一括処理</strong> - フレーム生成からGIF作成までを一括で実行します。</p>
            </div>
        </div>
        
        <div class="tab-content" id="random-comparison">
            <h2>ランダムデータとの比較</h2>
            
            <div class="form-group">
                <label for="initialSongs">初期曲数:</label>
                <input type="number" id="initialSongs" value="50" min="10" max="500">
            </div>
            
            <div class="form-group">
                <label for="frames">フレーム数:</label>
                <input type="number" id="frames" value="100" min="10" max="500">
            </div>
            
            <div class="form-group">
                <label for="distType">分布タイプ:</label>
                <select id="distType">
                    <option value="uniform">一様分布</option>
                    <option value="normal">正規分布（中心集中）</option>
                    <option value="biased">偏りのある分布（クラスター）</option>
                </select>
            </div>
            
            <div class="button-group">
                <button id="generateRandom">ランダムデータGIF生成</button>
                <button id="compareVisualizations">シミュレーションと比較</button>
            </div>
            
            <div class="progress-container">
                <div class="progress-bar" id="randomProgress"></div>
            </div>
            <div class="progress-text" id="randomStatus">準備完了</div>
            
            <div class="comparison-view">
                <div class="gif-container">
                    <h3>シミュレーション</h3>
                    <img id="simulationGIF" class="gif-display" src="" alt="シミュレーションGIF" style="display: none;">
                </div>
                <div class="gif-container">
                    <h3>ランダムデータ</h3>
                    <img id="randomGIF" class="gif-display" src="" alt="ランダムデータGIF" style="display: none;">
                </div>
            </div>
            
            <div class="info-text">
                <p><strong>パラメータ説明:</strong></p>
                <p><strong>初期曲数</strong> - ランダムデータで生成する曲の数。シミュレーションと比較可能な数に設定します。</p>
                <p><strong>フレーム数</strong> - アニメーションのフレーム数。多いほど滑らかになります。</p>
                <p><strong>分布タイプ</strong> - ランダムデータのジャンル分布タイプ。一様分布はランダム、正規分布は中心集中、偏りのある分布はクラスターを形成します。</p>
            </div>
        </div>
        
        <div class="tab-content" id="graphs">
            <h2>時系列グラフ分析</h2>
            
            <div class="button-group">
                <button id="loadGraphs">グラフを読み込む</button>
                <button id="exportGraphData">データエクスポート</button>
            </div>
            
            <div class="chart-container">
                <h3>人口推移</h3>
                <canvas id="populationChart"></canvas>
            </div>
            
            <div class="chart-container">
                <h3>曲数の変化</h3>
                <canvas id="songsChart"></canvas>
            </div>
            
            <div class="chart-container">
                <h3>ジャンル多様性</h3>
                <canvas id="diversityChart"></canvas>
            </div>
            
            <div class="chart-container">
                <h3>ジャンルクラスター数</h3>
                <canvas id="clustersChart"></canvas>
            </div>
            
            <div class="info-text">
                <p><strong>グラフ説明:</strong></p>
                <p><strong>人口推移</strong> - 時間の経過に伴うエージェント（全体、クリエイター、リスナー）の数の変化を示します。</p>
                <p><strong>曲数の変化</strong> - 時間の経過に伴って存在する曲の数の変化を示します。</p>
                <p><strong>ジャンル多様性</strong> - ジャンル空間内のユニークな位置の割合。1に近いほど多様性が高く、0に近いほど同質的です。</p>
                <p><strong>ジャンルクラスター数</strong> - ジャンル空間内で曲が集中している領域の数の変化を示します。</p>
            </div>
        </div>
    </div>

    <script>
        // タブ切り替え
        document.querySelectorAll('.tab').forEach(tab => {
            tab.addEventListener('click', () => {
                document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
                document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
                
                tab.classList.add('active');
                document.getElementById(tab.dataset.tab).classList.add('active');
            });
        });

        // スライダー値の表示
        const majorProbSlider = document.getElementById('majorProb');
        const majorProbValue = document.getElementById('majorProbValue');
        
        majorProbSlider.addEventListener('input', () => {
            majorProbValue.textContent = majorProbSlider.value;
        });

        // 進捗状況取得関数
        function updateProgress() {
            fetch('/api/get-progress')
                .then(response => response.json())
                .then(data => {
                    const progressBar = document.getElementById(getActiveTabId() + 'Progress');
                    const statusText = document.getElementById(getActiveTabId() + 'Status');
                    
                    progressBar.style.width = (data.progress * 100) + '%';
                    statusText.textContent = data.message;
                    
                    if (data.status === 'completed') {
                        // 完了時の処理
                        if (getActiveTabId() === 'visualization') {
                            document.getElementById('genreGIF').src = '/genre_evolution.gif?' + new Date().getTime();
                            document.getElementById('genreGIF').style.display = 'block';
                        } else if (getActiveTabId() === 'random') {
                            document.getElementById('randomGIF').src = '/random_evolution.gif?' + new Date().getTime();
                            document.getElementById('randomGIF').style.display = 'block';
                            
                            // 比較モードの場合、シミュレーションGIFも表示
                            if (document.getElementById('simulationGIF').src) {
                                document.getElementById('simulationGIF').style.display = 'block';
                            }
                        }
                    }
                    
                    // 未完了ならポーリング継続
                    if (data.status !== 'completed' && data.status !== 'error') {
                        setTimeout(updateProgress, 1000);
                    }
                })
                .catch(error => {
                    console.error('進捗取得エラー:', error);
                });
        }

        // アクティブなタブIDを取得
        function getActiveTabId() {
            const activeTab = document.querySelector('.tab.active');
            const tabId = activeTab.dataset.tab;
            
            if (tabId === 'simulation') return 'simulation';
            else if (tabId === 'visualization') return 'visualization';
            else if (tabId === 'random-comparison') return 'random';
        }

        // シミュレーション実行
        document.getElementById('runSimulation').addEventListener('click', () => {
            const majorProb = parseFloat(document.getElementById('majorProb').value);
            const initialAgents = parseInt(document.getElementById('initialAgents').value);
            const iterations = parseInt(document.getElementById('iterations').value);
            
            fetch('/api/run-simulation', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    majorProb,
                    initialAgents,
                    iterations
                })
            })
            .then(response => response.json())
            .then(data => {
                console.log('シミュレーション開始:', data);
                updateProgress();
            })
            .catch(error => {
                console.error('シミュレーションエラー:', error);
            });
        });

        // フレーム生成
        document.getElementById('generateFrames').addEventListener('click', () => {
            fetch('/api/generate-frames', {
                method: 'POST'
            })
            .then(response => response.json())
            .then(data => {
                console.log('フレーム生成開始:', data);
                updateProgress();
            })
            .catch(error => {
                console.error('フレーム生成エラー:', error);
            });
        });

        // GIF作成
        document.getElementById('createGIF').addEventListener('click', () => {
            fetch('/api/create-gif', {
                method: 'POST'
            })
            .then(response => response.json())
            .then(data => {
                console.log('GIF作成開始:', data);
                updateProgress();
            })
            .catch(error => {
                console.error('GIF作成エラー:', error);
            });
        });

        // 一括処理
        document.getElementById('runVisualization').addEventListener('click', () => {
            // まずフレーム生成
            fetch('/api/generate-frames', {
                method: 'POST'
            })
            .then(response => response.json())
            .then(data => {
                console.log('フレーム生成開始:', data);
                
                // フレーム生成が完了したらGIF作成
                const checkFrameCompletion = setInterval(() => {
                    fetch('/api/get-progress')
                        .then(response => response.json())
                        .then(progressData => {
                            if (progressData.status === 'completed') {
                                clearInterval(checkFrameCompletion);
                                
                                // GIF作成開始
                                fetch('/api/create-gif', {
                                    method: 'POST'
                                })
                                .then(response => response.json())
                                .then(gifData => {
                                    console.log('GIF作成開始:', gifData);
                                    updateProgress();
                                });
                            }
                        });
                }, 1000);
                
                updateProgress();
            })
            .catch(error => {
                console.error('可視化エラー:', error);
            });
        });

        // ランダムデータGIF生成
        document.getElementById('generateRandom').addEventListener('click', () => {
            const initialSongs = parseInt(document.getElementById('initialSongs').value);
            const frames = parseInt(document.getElementById('frames').value);
            const distType = document.getElementById('distType').value;
            
            fetch('/api/random-visualization', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    initialSongs,
                    frames,
                    distType
                })
            })
            .then(response => response.json())
            .then(data => {
                console.log('ランダムデータ生成開始:', data);
                updateProgress();
            })
            .catch(error => {
                console.error('ランダムデータ生成エラー:', error);
            });
        });

        // シミュレーションと比較
        document.getElementById('compareVisualizations').addEventListener('click', () => {
            // シミュレーションGIFが存在するか確認
            fetch('/genre_evolution.gif', { method: 'HEAD' })
                .then(response => {
                    if (response.ok) {
                        document.getElementById('simulationGIF').src = '/genre_evolution.gif?' + new Date().getTime();
                        document.getElementById('simulationGIF').style.display = 'block';
                        
                        // ランダムデータがなければ生成
                        fetch('/random_evolution.gif', { method: 'HEAD' })
                            .then(randomResponse => {
                                if (randomResponse.ok) {
                                    document.getElementById('randomGIF').src = '/random_evolution.gif?' + new Date().getTime();
                                    document.getElementById('randomGIF').style.display = 'block';
                                } else {
                                    // ランダムデータを生成
                                    document.getElementById('generateRandom').click();
                                }
                            });
                    } else {
                        alert('先にシミュレーションを実行し、可視化してください。');
                    }
                });
        });

        // グラフ作成機能
        let charts = {};
        
        document.getElementById('loadGraphs').addEventListener('click', () => {
            fetch('/api/time-series-data')
                .then(response => {
                    if (!response.ok) {
                        throw new Error('データの取得に失敗しました。先にシミュレーションを実行してください。');
                    }
                    return response.json();
                })
                .then(data => {
                    createCharts(data);
                })
                .catch(error => {
                    alert(error.message);
                    console.error('グラフデータ取得エラー:', error);
                });
        });
        
        document.getElementById('exportGraphData').addEventListener('click', () => {
            fetch('/api/time-series-data')
                .then(response => {
                    if (!response.ok) {
                        throw new Error('データの取得に失敗しました。先にシミュレーションを実行してください。');
                    }
                    return response.json();
                })
                .then(data => {
                    // JSONをダウンロード用にBlobに変換
                    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
                    const url = URL.createObjectURL(blob);
                    
                    // ダウンロードリンクを作成
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = 'musl_time_series_data.json';
                    document.body.appendChild(a);
                    a.click();
                    
                    // クリーンアップ
                    setTimeout(() => {
                        document.body.removeChild(a);
                        URL.revokeObjectURL(url);
                    }, 0);
                })
                .catch(error => {
                    alert(error.message);
                    console.error('データエクスポートエラー:', error);
                });
        });
        
        function createCharts(data) {
            // 既存のチャートがあれば破棄
            Object.values(charts).forEach(chart => chart.destroy && chart.destroy());
            charts = {};
            
            // 人口推移グラフ
            charts.population = new Chart(
                document.getElementById('populationChart').getContext('2d'),
                {
                    type: 'line',
                    data: {
                        labels: data.timeSteps,
                        datasets: [
                            {
                                label: '全エージェント',
                                data: data.population.total,
                                borderColor: 'rgb(54, 162, 235)',
                                backgroundColor: 'rgba(54, 162, 235, 0.1)',
                                tension: 0.1
                            },
                            {
                                label: 'クリエイター',
                                data: data.population.creators,
                                borderColor: 'rgb(255, 99, 132)',
                                backgroundColor: 'rgba(255, 99, 132, 0.1)',
                                tension: 0.1
                            },
                            {
                                label: 'リスナー',
                                data: data.population.listeners,
                                borderColor: 'rgb(75, 192, 192)',
                                backgroundColor: 'rgba(75, 192, 192, 0.1)',
                                tension: 0.1
                            }
                        ]
                    },
                    options: {
                        responsive: true,
                        plugins: {
                            title: {
                                display: true,
                                text: 'エージェント人口の推移'
                            }
                        },
                        scales: {
                            x: {
                                title: {
                                    display: true,
                                    text: 'イテレーション'
                                }
                            },
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: 'エージェント数'
                                }
                            }
                        }
                    }
                }
            );
            
            // 曲数の変化グラフ
            charts.songs = new Chart(
                document.getElementById('songsChart').getContext('2d'),
                {
                    type: 'line',
                    data: {
                        labels: data.timeSteps,
                        datasets: [
                            {
                                label: '曲数',
                                data: data.songs,
                                borderColor: 'rgb(153, 102, 255)',
                                backgroundColor: 'rgba(153, 102, 255, 0.1)',
                                tension: 0.1
                            }
                        ]
                    },
                    options: {
                        responsive: true,
                        plugins: {
                            title: {
                                display: true,
                                text: '曲数の推移'
                            }
                        },
                        scales: {
                            x: {
                                title: {
                                    display: true,
                                    text: 'イテレーション'
                                }
                            },
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: '曲数'
                                }
                            }
                        }
                    }
                }
            );
            
            // ジャンル多様性グラフ
            charts.diversity = new Chart(
                document.getElementById('diversityChart').getContext('2d'),
                {
                    type: 'line',
                    data: {
                        labels: data.timeSteps,
                        datasets: [
                            {
                                label: 'ジャンル多様性指数',
                                data: data.genres.diversity,
                                borderColor: 'rgb(255, 159, 64)',
                                backgroundColor: 'rgba(255, 159, 64, 0.1)',
                                tension: 0.1
                            }
                        ]
                    },
                    options: {
                        responsive: true,
                        plugins: {
                            title: {
                                display: true,
                                text: 'ジャンル多様性の推移'
                            }
                        },
                        scales: {
                            x: {
                                title: {
                                    display: true,
                                    text: 'イテレーション'
                                }
                            },
                            y: {
                                beginAtZero: true,
                                max: 1,
                                title: {
                                    display: true,
                                    text: '多様性指数 (0-1)'
                                }
                            }
                        }
                    }
                }
            );
            
            // ジャンルクラスター数グラフ
            charts.clusters = new Chart(
                document.getElementById('clustersChart').getContext('2d'),
                {
                    type: 'line',
                    data: {
                        labels: data.timeSteps,
                        datasets: [
                            {
                                label: 'クラスター数',
                                data: data.genres.clusters,
                                borderColor: 'rgb(201, 203, 207)',
                                backgroundColor: 'rgba(201, 203, 207, 0.1)',
                                tension: 0.1
                            }
                        ]
                    },
                    options: {
                        responsive: true,
                        plugins: {
                            title: {
                                display: true,
                                text: 'ジャンルクラスター数の推移'
                            }
                        },
                        scales: {
                            x: {
                                title: {
                                    display: true,
                                    text: 'イテレーション'
                                }
                            },
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: 'クラスター数'
                                }
                            }
                        }
                    }
                }
            );
        }
    </script>
</body>
</html>
`

	// インデックスファイルの保存
	indexPath := filepath.Join(staticDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(htmlContent), 0644); err != nil {
		log.Fatalf("HTMLファイルの作成に失敗しました: %v", err)
	}
}
