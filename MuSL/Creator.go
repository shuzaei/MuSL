package MuSL

type Song struct {
	id      int
	genre   []float64
	creator *Agent
}

type Creator struct {
	innovation_rate      float64
	memory               []*Song
	creation_probability float64
	creation_cost        const64
}

func (c *Creator) Create(agents *[]*Agent) {
	// あとで実装
}
