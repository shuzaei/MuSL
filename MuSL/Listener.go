package MuSL

type Listener struct {
	novelty_preference    float64
	memory                []*Song
	incoming_songs        []*Song
	listening_probability float64
	evaluation_cost       const64
}

func (l *Listener) Listen(agent *[]*Agent) {
	// あとで実装
}
