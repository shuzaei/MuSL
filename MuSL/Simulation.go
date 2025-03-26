package MuSL

// シミュレーションの骨格
type Simulation struct {
	agents               []*Agent
	n_iter               int
	ga_params            *GAParams
	default_agent_params *Agent
}

// 新しいシミュレーションを作成
func MakeNewSimulation(n_agents, n_iter int, ga_params *GAParams, default_agent_params *Agent) *Simulation {
	sim := &Simulation{
		agents:               make([]*Agent, n_agents),
		n_iter:               n_iter,
		ga_params:            ga_params,
		default_agent_params: default_agent_params,
	}

	// エージェントを作成
	for i := range n_agents {
		sim.agents[i] = MakeRandomAgentFromParams(GetNewID(), default_agent_params)
	}

	return sim
}

// シミュレーションを実行
func (s *Simulation) Run() {
	for range s.n_iter {
		// new_agents にエージェントをコピー
		// その際、エネルギーが 0 以下のエージェントを削除
		new_agents := make([]*Agent, 0)
		for _, agent := range s.agents {
			if agent.energy > 0 {
				new_agents = append(new_agents, agent)
			}
		}

		// エージェントを実行
		for _, agent := range new_agents {
			agent.Run(&new_agents, s.ga_params, s.default_agent_params)
		}

		// エージェントを保存
		s.agents = new_agents
	}
}

// シミュレーションの結果を返す
func (s *Simulation) GetResult() error {
	return nil
}
