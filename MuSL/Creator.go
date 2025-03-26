package MuSL

type Song struct {
	id      int
	genre   []float64
	creator int
}

type Creator struct {
	innovation_rate      float64
	memory               []*Song
	creation_probability float64
	creation_cost        const64
}
