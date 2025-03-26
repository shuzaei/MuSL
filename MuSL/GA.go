package MuSL

import (
	"math"
	"math/rand/v2"
)

// 遺伝子型。Agent 固有の遺伝子に加えて Creator, Listener, Organizer の遺伝子もつなげたものになる。
// 役割は各役割のオンオフを 0.0 / 1.0 に変換し、帰ってきた値が 0.5 以上ならその役割を持つと判断する。
// 役割がないと何もしなくて問題なので、その場合は失敗としてエラーを返す。（というのを Agent.go で実装している）

type Evolvable interface {
	ToGene() []float64
	FromGene([]float64) error
}

type GAParams struct {
	mutation_rate     float64 // 例: 0.1
	mutation_strength float64 // 例: 0.05
}

func MakeGAParams(mutation_rate, mutation_strength float64) *GAParams {
	return &GAParams{
		mutation_rate:     mutation_rate,
		mutation_strength: mutation_strength,
	}
}

func ReproduceGA[T Evolvable](p1, p2 T, params *GAParams, default_params T, copy_func func(T) T) (T, error) {
	g1 := p1.ToGene()
	g2 := p2.ToGene()

	childGene := CrossoverAndMutate(g1, g2, params)

	child := copy_func(default_params)

	err := child.FromGene(childGene)
	return child, err
}

func CrossoverAndMutate(g1, g2 []float64, params *GAParams) []float64 {
	childGene := make([]float64, len(g1))
	for i := range g1 {
		if rand.Float64() < 0.5 {
			childGene[i] = g1[i]
		} else {
			childGene[i] = g2[i]
		}

		if rand.Float64() < params.mutation_rate {
			childGene[i] += params.mutation_strength * (rand.Float64()*2.0 - 1.0)
		}

		childGene[i] = math.Max(0.0, math.Min(1.0, childGene[i]))
	}
	return childGene
}
