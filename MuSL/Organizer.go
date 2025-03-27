package MuSL

import (
	"math/rand/v2"
	"sort"
)

type Event struct {
	event_type        string
	creator_pool      []*Song
	listener_pool     []*Agent
	evaluation_pool   map[*Song][]float64
	evaluation_reward map[*Song]float64
}

type Organizer struct {
	major_probability   const64
	created_events      []*Event
	event_probability   float64
	organization_cost   const64
	organization_reward const64

	// イベント生成用のパラメータ
	// メジャーイベント
	major_listener_ratio const64
	major_creator_ratio  const64
	major_song_ratio     const64
	major_winner_ratio   const64 // 上位何%に報酬を与えるか
	major_reward_ratio   const64 // 上位に与える報酬の割合

	// マイナーイベント
	minor_listener_ratio const64
	minor_creator_ratio  const64
	minor_song_ratio     const64
	minor_reward_ratio   const64 // そのまま報酬を与える割合
}

func (o *Organizer) Organize(agents *[]*Agent, me *Agent, summery *Summery) {
	// 前回のイベントの報酬を支払う
	for _, event := range o.created_events {
		// イベントの報酬を支払う

		if event.event_type == "major" {

			// 合計報酬を計算
			reward_sum := 0.0
			for _, reward := range event.evaluation_reward {
				reward_sum += reward
			}

			// 最初に中抜きを行う
			fee := reward_sum * float64(o.organization_reward)

			reward_sum -= fee
			me.energy += reward_sum

			// 曲を平均評価値でソート
			// 平均評価値の計算
			type SongEvaluation struct {
				song       *Song
				evaluation float64
			}

			song_evaluations := make([]SongEvaluation, 0)
			for song, evaluations := range event.evaluation_pool {
				sum := 0.0
				for _, evaluation := range evaluations {
					sum += evaluation
				}
				average := sum / float64(len(evaluations))
				song_evaluations = append(song_evaluations, SongEvaluation{song, average})
			}

			// 平均評価値でソート
			sort.Slice(song_evaluations, func(i, j int) bool {
				return song_evaluations[i].evaluation > song_evaluations[j].evaluation
			})

			// 上位の曲に報酬を与える
			num_winners := int(float64(len(song_evaluations)) * float64(o.major_winner_ratio))
			bonus := reward_sum * float64(o.major_reward_ratio) / float64(num_winners)

			for _, song_evaluation := range song_evaluations[:num_winners] {
				song := song_evaluation.song
				song.creator.energy += bonus
			}

			// 全ての曲に報酬を与える
			each_reward := reward_sum * (1.0 - float64(o.major_reward_ratio)) / float64(len(song_evaluations))
			for _, song_evaluation := range song_evaluations {
				song := song_evaluation.song
				song.creator.energy += each_reward
			}
		} else {
			// マイナーイベント
			// マイナーイベントでは、最初に一定割合の報酬を還元
			reward_sum := 0.0

			for song, reward := range event.evaluation_reward {
				// 中抜き
				fee := reward * float64(o.organization_reward)
				reward -= fee
				me.energy += reward

				// 一定割合を還元
				reward_return := reward * float64(o.minor_reward_ratio)
				song.creator.energy += reward_return
				reward_sum += reward - reward_return
			}

			// 全ての曲に報酬を与える
			each_reward := reward_sum / float64(len(event.evaluation_reward))
			for song := range event.evaluation_reward {
				song.creator.energy += each_reward
			}
		}
	}

	// イベントをすべて削除
	o.created_events = make([]*Event, 0)

	if rand.Float64() < o.event_probability {
		// イベントを生成
		event_type := ""
		creator_pool := make([]*Song, 0)
		listener_pool := make([]*Agent, 0)
		evaluation_pool := make(map[*Song][]float64)
		evaluation_reward := make(map[*Song]float64)

		creators := make([]*Agent, 0)

		if rand.Float64() < float64(o.major_probability) {
			// メジャーイベント
			event_type = "major"
			for _, agent := range *agents {
				if agent.role[0] && rand.Float64() < float64(o.major_creator_ratio) {
					creators = append(creators, agent)
				}
				if agent.role[1] && rand.Float64() < float64(o.major_listener_ratio) {
					listener_pool = append(listener_pool, agent)
				}
			}

			for _, creator := range creators {
				for _, song := range creator.creator.memory {
					if rand.Float64() < float64(o.major_song_ratio) {
						creator_pool = append(creator_pool, song)
					}
				}
			}
		} else {
			// マイナーイベント
			event_type = "minor"
			for _, agent := range *agents {
				if agent.role[0] && rand.Float64() < float64(o.minor_creator_ratio) {
					creators = append(creators, agent)
				}
				if agent.role[1] && rand.Float64() < float64(o.minor_listener_ratio) {
					listener_pool = append(listener_pool, agent)
				}
			}

			for _, creator := range creators {
				for _, song := range creator.creator.memory {
					if rand.Float64() < float64(o.minor_song_ratio) {
						creator_pool = append(creator_pool, song)
					}
				}
			}
		}

		// イベントを生成
		event := &Event{
			event_type:        event_type,
			creator_pool:      creator_pool,
			listener_pool:     listener_pool,
			evaluation_pool:   evaluation_pool,
			evaluation_reward: evaluation_reward,
		}

		o.created_events = append(o.created_events, event)

		// 開催コストを支払う
		me.energy -= float64(o.organization_cost)

		// イベントの報酬を設定
		for _, song := range creator_pool {
			event.evaluation_pool[song] = make([]float64, 0)
			event.evaluation_reward[song] = 0.0
		}

		// 集計 (III)
		summery.num_event_all++
		summery.num_event_this++

		// 評価は次のイテレーションまでに集められる
	}
}
