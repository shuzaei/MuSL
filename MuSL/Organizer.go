package MuSL

type Event struct {
	event_type        string
	creator_pool      []*Song
	listener_pool     []*Listener
	num_winners       int
	reward_ratio      float64
	evaluation_pool   map[*Song][]float64
	evaluation_reward map[*Song]float64
}

type Organizer struct {
	major_probability   const64
	created_events      []*Event
	event_probability   float64
	organization_cost   const64
	organization_reward const64
}
