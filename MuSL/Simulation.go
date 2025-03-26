package MuSL

// シミュレーションの骨格
type Simulation struct {
	history              [][]*Agent
	n_iter               int
	ga_params            *GAParams
	default_agent_params *Agent
}

// 新しいシミュレーションを作成
func MakeNewSimulation(n_agents, n_iter int, ga_params *GAParams, default_agent_params *Agent) *Simulation {
	sim := &Simulation{
		history:              make([][]*Agent, n_iter+1),
		n_iter:               n_iter,
		ga_params:            ga_params,
		default_agent_params: default_agent_params,
	}

	// エージェントを作成
	agents := make([]*Agent, n_agents)
	for i := 0; i < n_agents; i++ {
		agents[i] = MakeRandomAgentFromParams(GetNewID(), default_agent_params)
	}

	sim.history[0] = agents

	return sim
}

// シミュレーションを実行
func (s *Simulation) Run() {
	for i := 0; i < s.n_iter; i++ {
		// history[i] のエージェントを使って history[i+1] を作成
		// その際、エネルギーが 0 以下のエージェントを削除
		agents := make([]*Agent, 0)
		for _, agent := range s.history[i] {
			if agent.energy > 0 {
				agents = append(agents, agent)
			}
		}

		// エージェントを実行
		for _, agent := range agents {
			agent.Run(&agents, s.ga_params, s.default_agent_params)
		}

		// エージェントを保存
		s.history[i+1] = agents
	}
}

// シミュレーションの結果を返す
func (s *Simulation) GetResult() [][]*Agent {
	return s.history
}
