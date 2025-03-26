package MuSL

import (
	"math"
	"math/rand"
)

type Song struct {
	genre   []float64
	creator *Agent
}

type Creator struct {
	innovation_rate      float64
	memory               []*Song
	creation_probability float64
	creation_cost        const64
}

func (c *Creator) Create(agents *[]*Agent, me *Agent) {
	if rand.Float64() < c.creation_probability {
		// 曲を生成
		genre := make([]float64, 2)

		// innovation rate に従ってジャンルを生成
		// memory からランダムに選んで突然変異

		// memory が空の場合はランダムに生成
		if len(c.memory) == 0 {
			for i := 0; i < len(genre); i++ {
				genre[i] = rand.Float64()
			}
		} else {
			random_index := rand.Intn(len(c.memory))
			for i := 0; i < len(genre); i++ {
				genre[i] = c.memory[random_index].genre[i] + rand.Float64()*c.innovation_rate

				// 0 以上 1 未満に収める
				genre[i] = math.Max(0.0, math.Min(1.0, genre[i]))
			}
		}

		// 曲を生成して memory に追加
		song := &Song{
			genre:   genre,
			creator: me,
		}
		c.memory = append(c.memory, song)

		// エネルギーを消費
		me.energy -= float64(c.creation_cost)
	}
}
