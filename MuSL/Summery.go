package MuSL

// シミュレーションのサマリー
type Summery struct {
	// [*] は、イテレーションの最後に Calculate で計算するもの
	num_population         int     // [*] 人数 (B)
	num_creaters           int     // [*] 作成者の人数 (B)
	num_listeners          int     // [*] 聴取者の人数 (B)
	num_organizers         int     // [*] 運営者の人数 (B)
	num_song_all           int     //     いままで作成された楽曲の総数 (E)
	num_song_this          int     //     そのイテレーションで作成された楽曲の総数 (E)
	num_song_now           int     // [*] 現在残っているエージェントの楽曲の総数 (E)
	num_evaluation_all     int     //     いままで行われた評価の総数 (E)
	num_evaluation_this    int     //     そのイテレーションで行われた評価の総数 (E)
	num_event_all          int     //     いままで開催されたイベントの総数 (E)
	num_event_this         int     //     そのイテレーションで開催されたイベントの総数 (E)
	avg_innovation         float64 // [*] 作成者の新規性の平均 (C)
	avg_novelty_preference float64 // [*] 聴取者の新規性好みの平均 (C)
	sum_evaluation         float64 //     そのイテレーションで行われた評価の合計 (G)
	avg_evaluation         float64 // [*] そのイテレーションで行われた評価の平均 (G)
	total_energy           float64 // [*] エネルギーの総量 (D)
	energy_creators        float64 // [*] 作成者のエネルギーの総量 (F)
	energy_listeners       float64 // [*] 聴取者のエネルギーの総量 (F)
	energy_organizers      float64 // [*] 運営者のエネルギーの総量 (F)
	all_genres             [][]float64
}

type PublicSummery struct {
	NumPopulation        int     `json:"num_population"`
	NumCreaters          int     `json:"num_creaters"`
	NumListeners         int     `json:"num_listeners"`
	NumOrganizers        int     `json:"num_organizers"`
	NumSongAll           int     `json:"num_song_all"`
	NumSongThis          int     `json:"num_song_this"`
	NumSongNow           int     `json:"num_song_now"`
	NumEvaluationAll     int     `json:"num_evaluation_all"`
	NumEvaluationThis    int     `json:"num_evaluation_this"`
	NumEventAll          int     `json:"num_event_all"`
	NumEventThis         int     `json:"num_event_this"`
	AvgInnovation        float64 `json:"avg_innovation"`
	AvgNoveltyPreference float64 `json:"avg_novelty_preference"`
	SumEvaluation        float64 `json:"sum_evaluation"`
	AvgEvaluation        float64 `json:"avg_evaluation"`
	TotalEnergy          float64 `json:"total_energy"`
	EnergyCreators       float64 `json:"energy_creators"`
	EnergyListeners      float64 `json:"energy_listeners"`
	EnergyOrganizers     float64 `json:"energy_organizers"`
	AllGenres            [][]float64
}

func MakeNewSummery() *Summery {
	return &Summery{
		num_population:         0,
		num_creaters:           0,
		num_listeners:          0,
		num_organizers:         0,
		num_song_all:           0,
		num_song_this:          0,
		num_song_now:           0,
		num_evaluation_all:     0,
		num_evaluation_this:    0,
		num_event_all:          0,
		num_event_this:         0,
		avg_innovation:         0,
		avg_novelty_preference: 0,
		sum_evaluation:         0,
		avg_evaluation:         0,
		total_energy:           0,
		energy_creators:        0,
		energy_listeners:       0,
		energy_organizers:      0,
		all_genres:             [][]float64{},
	}
}

func MakeNewSummeryFromSummery(s *Summery) *Summery {
	return &Summery{
		num_population:         0,                    // 1-1 再計算
		num_creaters:           0,                    // 1-2 再計算
		num_listeners:          0,                    // 1-3 再計算
		num_organizers:         0,                    // 1-4 再計算
		num_song_all:           s.num_song_all,       // 加算 (I)
		num_song_this:          0,                    // リセットして集計 (I)
		num_song_now:           0,                    // 2 再計算
		num_evaluation_all:     s.num_evaluation_all, // 加算 (II)
		num_evaluation_this:    0,                    // リセットして集計 (II)
		num_event_all:          s.num_event_all,      // 加算 (III)
		num_event_this:         0,                    // リセットして集計 (III)
		avg_innovation:         0,                    // 3-1 再計算
		avg_novelty_preference: 0,                    // 3-2 再計算
		sum_evaluation:         0,                    // リセットして集計 (IV)
		avg_evaluation:         0,                    // 4 再計算
		total_energy:           0,                    // 5-1 再計算
		energy_creators:        0,                    // 5-2 再計算
		energy_listeners:       0,                    // 5-3 再計算
		energy_organizers:      0,                    // 5-4 再計算
		all_genres:             [][]float64{},        // 6 再取得
	}
}

func (s *Summery) Publish() *PublicSummery {
	return &PublicSummery{
		NumPopulation:        s.num_population,
		NumCreaters:          s.num_creaters,
		NumListeners:         s.num_listeners,
		NumOrganizers:        s.num_organizers,
		NumSongAll:           s.num_song_all,
		NumSongThis:          s.num_song_this,
		NumSongNow:           s.num_song_now,
		NumEvaluationAll:     s.num_evaluation_all,
		NumEvaluationThis:    s.num_evaluation_this,
		NumEventAll:          s.num_event_all,
		NumEventThis:         s.num_event_this,
		AvgInnovation:        s.avg_innovation,
		AvgNoveltyPreference: s.avg_novelty_preference,
		SumEvaluation:        s.sum_evaluation,
		AvgEvaluation:        s.avg_evaluation,
		TotalEnergy:          s.total_energy,
		EnergyCreators:       s.energy_creators,
		EnergyListeners:      s.energy_listeners,
		EnergyOrganizers:     s.energy_organizers,
		AllGenres:            s.all_genres,
	}
}

func PublishAllSummery(summery []*Summery) []*PublicSummery {
	ret := make([]*PublicSummery, len(summery))
	for i, s := range summery {
		ret[i] = s.Publish()
	}
	return ret
}

func (s *Summery) Calculate(agents []*Agent) {

	// s がすでにリセットされているものとして、各項目を計算
	for _, agent := range agents {
		// 生きているエージェントのみ
		if agent.energy <= 0 {
			continue
		}

		s.num_population++             // 1-1
		s.total_energy += agent.energy // 5-1

		// エージェントの役割ごとに集計
		if agent.role[0] {
			s.num_creaters++                  // 1-2
			s.energy_creators += agent.energy // 5-2

			// innovation
			s.avg_innovation += agent.creator.innovation_rate // 3-1
		}
		if agent.role[1] {
			s.num_listeners++                  // 1-3
			s.energy_listeners += agent.energy // 5-3

			// novelty preference
			s.avg_novelty_preference += agent.listener.novelty_preference // 3-2
		}
		if agent.role[2] {
			s.num_organizers++                  // 1-4
			s.energy_organizers += agent.energy // 5-4
		}

		// 残っている楽曲の数
		s.num_song_now += len(agent.creator.memory) // 2
		for _, song := range agent.creator.memory {
			s.all_genres = append(s.all_genres, song.genre) // 6
		}
	}

	// 最後に平均を計算
	if s.num_creaters > 0 {
		s.avg_innovation /= float64(s.num_creaters) // 3-1
	}
	if s.num_listeners > 0 {
		s.avg_novelty_preference /= float64(s.num_listeners) // 3-2
	}
	if s.num_evaluation_this > 0 {
		s.avg_evaluation = s.sum_evaluation / float64(s.num_evaluation_this) // 4
	}
}
