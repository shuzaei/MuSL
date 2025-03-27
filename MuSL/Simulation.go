package MuSL

// シミュレーションの骨格
type Simulation struct {
	agents               []*Agent
	n_iter               int
	ga_params            *GAParams
	default_agent_params *Agent
	summery              []*Summery
}

// 新しいシミュレーションを作成
func MakeNewSimulation(n_agents, n_iter int, ga_params *GAParams, default_agent_params *Agent) *Simulation {
	sim := &Simulation{
		agents:               make([]*Agent, n_agents),
		n_iter:               n_iter,
		ga_params:            ga_params,
		default_agent_params: default_agent_params,
		summery:              make([]*Summery, n_iter+1),
	}

	// エージェントを作成
	for i := range n_agents {
		sim.agents[i] = MakeRandomAgentFromParams(GetNewID(), default_agent_params)
	}

	// サマリーを作成
	sim.summery[0] = MakeNewSummery()

	return sim
}

// シミュレーションを実行
func (s *Simulation) Run() {
	for i := range s.n_iter {
		// new_agents にエージェントをコピー
		// その際、エネルギーが 0 以下のエージェントを削除
		new_agents := make([]*Agent, 0)
		for _, agent := range s.agents {
			if agent.energy > 0 {
				new_agents = append(new_agents, agent)
			}
		}

		// サマリーを作成
		s.summery[i+1] = CopySummery(s.summery[i])

		// 新しく生まれたエージェントを入れるプール
		new_born_pool := make([]*Agent, 0)

		// エージェントを実行
		for _, agent := range new_agents {
			agent.Run(&new_agents, &new_born_pool,
				s.ga_params, s.default_agent_params, s.summery[i+1])
		}

		// エージェントを保存
		s.agents = append(new_agents, new_born_pool...)

		// サマリーを更新
		s.summery[i+1].Calculate(s.agents)
	}
}

// シミュレーションの結果を返す
func (s *Simulation) GetSummery() []*PublicSummery {
	return PublishAllSummery(s.summery)
}
