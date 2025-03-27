package main

import (
	"MuSL/MuSL"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	// コマンドライン引数でメジャーイベントの確率および出力先を指定
	var major_probability float64
	var output_file string

	flag.Float64Var(&major_probability, "major_probability", 0.5, "Probability of major events (default: 0.5)")
	flag.StringVar(&output_file, "output_file", "output.json", "Output file name (default: output.json)")
	flag.Parse()

	if major_probability < 0 || major_probability > 1 {
		fmt.Println("Invalid major_probability: must be between 0 and 1")
		return
	}

	// シミュレーションのパラメータ
	n_agents := 100
	n_iter := 100

	ga_params := MuSL.MakeGAParams(
		0.1,  // mutation_rate
		0.05, // mutation_strength
	)

	// [*] = Experiment-specific
	default_agent_params := MuSL.MakeNewAgent(
		-1,                       //     id
		[]bool{true, true, true}, //     role
		100.0,                    //     energy
		100.0,                    // [*] default_energy
		0.0,                      // [*] elimination_threshold
		0.5,                      //     reproduction_probability

		// creator
		0.5,                   //     innovation_rate
		make([]*MuSL.Song, 0), //     memory
		0.5,                   //     creation_probability
		2.0,                   // [*] creation_cost

		// listener
		0.5,                    //     novelty_preference
		make([]*MuSL.Song, 0),  //     memory
		make([]*MuSL.Song, 0),  //     incoming_songs
		make([]*MuSL.Event, 0), //     song_events
		0.5,                    //     listening_probability
		2.0,                    // [*] evaluation_cost

		// organizer
		MuSL.Const64(major_probability), // [*] major_probability
		make([]*MuSL.Event, 0),          //     created_events
		0.5,                             //     event_probability
		0.5,                             // [*] organization_cost
		2.0,                             // [*] organization_reward

		// イベント生成用のパラメータ
		// メジャーイベント
		0.5, // [*] major_listener_ratio
		0.5, // [*] major_creator_ratio
		0.1, // [*] major_song_ratio
		0.5, // [*] major_winner_ratio
		0.5, // [*] major_reward_ratio
		0.1, // [*] major_recommendation_ratio

		// マイナーイベント
		0.1, // [*] minor_listener_ratio
		0.1, // [*] minor_creator_ratio
		0.5, // [*] minor_song_ratio
		0.5, // [*] minor_reward_ratio
		0.1, // [*] minor_recommendation_ratio
	)

	sim := MuSL.MakeNewSimulation(n_agents, n_iter, ga_params, default_agent_params)
	sim.Run()

	summery := sim.GetSummery() // []*PublicSummery

	// サマリーを json に変換
	json, err := json.MarshalIndent(summery, "", "  ")
	if err != nil {
		fmt.Println("Error converting summery to JSON:", err)
		return
	}

	// ファイルに書き込み
	file, err := os.Create(output_file)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	file.Write(json)
}
