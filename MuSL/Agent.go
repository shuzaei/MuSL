package MuSL

import (
	"math/rand/v2"
	"strconv"
)

type const64 float64 // 実験時に確定する定数

type Agent struct {
	id                       int
	role                     []bool // [creator, listener, organizer]
	energy                   float64
	default_energy           const64
	elimination_threshold    const64
	reproduction_probability float64

	creator   *Creator
	listener  *Listener
	organizer *Organizer
}

func MakeNewAgent(
	id int,
	role []bool, // -------------------- Gene
	energy float64, // ----------------- 動的に変化
	default_energy const64, // --------- 実験定数
	elimination_threshold const64, // -- 実験定数
	reproduction_probability float64, // Gene

	// creator
	innovation_rate float64, // -------- Gene
	memory_c []*Song, // --------------- 動的に変化
	creation_probability float64, // --- Gene
	creation_cost const64, // ---------- 実験定数

	// listener
	novelty_preference float64, // ----- Gene
	memory_l []*Song, //---------------- 動的に変化
	incoming_songs []*Song, // --------- 動的に変化
	listening_probability float64, // -- Gene
	evaluation_cost const64, // -------- 実験定数

	// organizer
	major_probability const64, // ------ 実験定数
	created_events []*Event, // -------- 動的に変化
	event_probability float64, // ------ Gene
	organization_cost const64, // ------ 実験定数
	organization_reward const64, // ---- 実験定数
) *Agent {
	return &Agent{
		id:                       id,
		role:                     role,
		energy:                   float64(default_energy),
		default_energy:           default_energy,
		elimination_threshold:    elimination_threshold,
		reproduction_probability: reproduction_probability,
		creator:                  &Creator{innovation_rate, memory_c, creation_probability, creation_cost},
		listener:                 &Listener{novelty_preference, memory_l, incoming_songs, listening_probability, evaluation_cost},
		organizer:                &Organizer{major_probability, created_events, event_probability, organization_cost, organization_reward},
	}
}

func CopyAgent(a *Agent) *Agent {
	return MakeNewAgent(
		a.id,
		a.role,
		a.energy,
		a.default_energy,
		a.elimination_threshold,
		a.reproduction_probability,

		// creator
		a.creator.innovation_rate,
		a.creator.memory,
		a.creator.creation_probability,
		a.creator.creation_cost,

		// listener
		a.listener.novelty_preference,
		a.listener.memory,
		a.listener.incoming_songs,
		a.listener.listening_probability,
		a.listener.evaluation_cost,

		// organizer
		a.organizer.major_probability,
		a.organizer.created_events,
		a.organizer.event_probability,
		a.organizer.organization_cost,
		a.organizer.organization_reward,
	)
}

// 実験定数を受け取り、動的に変化するパラメータを初期化し、Gene をランダムで生成する
func MakeRandomAgentFromParams(
	id int,
	default_params *Agent) *Agent {
	role := []bool{
		rand.Float64() < 0.5,
		rand.Float64() < 0.5,
		rand.Float64() < 0.5,
	}

	return MakeNewAgent(
		id,
		role,
		float64(default_params.default_energy),
		default_params.default_energy,
		default_params.elimination_threshold,
		rand.Float64(),

		// creator
		rand.Float64(),
		make([]*Song, 0),
		rand.Float64(),
		default_params.creator.creation_cost,

		// listener
		rand.Float64(),
		make([]*Song, 0),
		make([]*Song, 0),
		rand.Float64(),
		default_params.listener.evaluation_cost,

		// organizer
		default_params.organizer.major_probability,
		make([]*Event, 0),
		rand.Float64(),
		default_params.organizer.organization_cost,
		default_params.organizer.organization_reward,
	)
}

func (a *Agent) Run(agents *[]*Agent, gaParams *GAParams, default_agent_params *Agent) {

	// リスナー
	if a.role[1] {
		a.listener.Listen(agents)
	}

	// クリエイター
	if a.role[0] {
		a.creator.Create(agents)
	}

	// オーガナイザー
	if a.role[2] {
		a.organizer.Organize(agents)
	}

	// 再生産
	a.Reproduce(agents, gaParams, default_agent_params)
}

func (a *Agent) Reproduce(agents *[]*Agent, gaParams *GAParams, default_agent_params *Agent) {
	// もしエネルギーが default_energy/2 以上なら、reproduction_probability の確率で子供を作る
	if a.energy < float64(a.default_energy)/2 {
		return
	}

	if rand.Float64() < a.reproduction_probability {
		// default_energy/2 以上の agent を探してランダムに選ぶ
		spouse_candidates := make([]*Agent, 0)
		for _, agent := range *agents {
			if agent.energy >= float64(a.default_energy)/2 {
				spouse_candidates = append(spouse_candidates, agent)
			}
		}

		if len(spouse_candidates) == 0 {
			return
		}

		spouse := spouse_candidates[rand.IntN(len(spouse_candidates))]
		child, err := ReproduceGA(a, spouse, gaParams, default_agent_params, CopyAgent)
		if err == nil {
			*agents = append(*agents, child)
		}

		// たまに失敗することもあるが、失敗してもエネルギーは減らす
		a.energy -= float64(a.default_energy) / 2
		spouse.energy -= float64(a.default_energy) / 2
	}
}

func (a *Agent) ToGene() []float64 {
	gene := make([]float64, 0)

	// role
	for _, r := range a.role {
		if r {
			gene = append(gene, 1.0)
		} else {
			gene = append(gene, 0.0)
		}
	}

	// reproduction_probability
	gene = append(gene, a.reproduction_probability)

	// creator
	// innovation_rate
	gene = append(gene, a.creator.innovation_rate)
	// creation_probability
	gene = append(gene, a.creator.creation_probability)

	// listener
	// novelty_preference
	gene = append(gene, a.listener.novelty_preference)
	// listening_probability
	gene = append(gene, a.listener.listening_probability)

	// organizer
	// event_probability
	gene = append(gene, a.organizer.event_probability)

	return gene
}

func (a *Agent) FromGene(gene []float64) error {
	if len(gene) != 9 {
		return &GeneLengthError{len(gene)}
	}

	// role
	for i := 0; i < 3; i++ {
		if gene[i] > 0.5 {
			a.role[i] = true
		} else {
			a.role[i] = false
		}
	}

	// role がすべて false ならエラー
	if !a.role[0] && !a.role[1] && !a.role[2] {
		return &NoRoleError{}
	}

	// reproduction_probability
	a.reproduction_probability = gene[3]

	// creator
	// innovation_rate
	a.creator.innovation_rate = gene[4]
	// creation_probability
	a.creator.creation_probability = gene[5]

	// listener
	// novelty_preference
	a.listener.novelty_preference = gene[6]
	// listening_probability
	a.listener.listening_probability = gene[7]

	// organizer
	// event_probability
	a.organizer.event_probability = gene[8]

	return nil
}

type GeneLengthError struct {
	length int
}

func (e *GeneLengthError) Error() string {
	return "Gene length is not 9: " + strconv.Itoa(e.length)
}

type NoRoleError struct{}

func (e *NoRoleError) Error() string {
	return "All roles are false"
}
