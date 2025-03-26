package main

import (
	"MuSL/MuSL"
	"fmt"
)

func main() {
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
		0.5,                   // [*] creation_cost

		// listener
		0.5,                   //     novelty_preference
		make([]*MuSL.Song, 0), //     memory
		make([]*MuSL.Song, 0), //     incoming_songs
		0.5,                   //     listening_probability
		0.5,                   // [*] evaluation_cost

		// organizer
		0.5,                    // [*] major_probability
		make([]*MuSL.Event, 0), //     created_events
		0.5,                    //     event_probability
		0.5,                    // [*] organization_cost
		0.5,                    // [*] organization_reward
	)

	sim := MuSL.MakeNewSimulation(n_agents, n_iter, ga_params, default_agent_params)
	sim.Run()

	result := sim.GetResult()
	// result を使って何かしたいが、まだ出力用 public method がないので、とりあえずアドレスを出力
	fmt.Println(result)
}
