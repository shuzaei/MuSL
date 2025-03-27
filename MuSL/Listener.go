package MuSL

import (
	"math"
	"math/rand/v2"
)

type Listener struct {
	novelty_preference    float64
	memory                []*Song
	incoming_songs        []*Song
	song_events           []*Event
	listening_probability float64
	evaluation_cost       const64
}

func (l *Listener) Listen(agents *[]*Agent, me *Agent, summery *Summery) {
	if len(l.incoming_songs) == 0 {
		return
	}

	// 順番に聴く
	for i, song := range l.incoming_songs {
		// 聴くかどうか
		if rand.Float64() < l.listening_probability {
			// 評価
			// 最も近い曲を探す
			min_distance := 1.0
			for _, memory_song := range l.memory {
				distance := 0.0
				for i := range song.genre {
					distance += (song.genre[i] - memory_song.genre[i]) * (song.genre[i] - memory_song.genre[i])
				}
				distance = distance / float64(len(song.genre))
				if distance < min_distance {
					min_distance = distance
				}
			}

			// novelty preference によって評価
			evaluation := 1 - math.Abs(min_distance-l.novelty_preference)/math.Sqrt(2) // 最大距離が sqrt(2) なので

			// 評価をイベントに記録し、エネルギーに加算
			l.song_events[i].evaluation_pool[song] = append(l.song_events[i].evaluation_pool[song], evaluation)
			me.energy += evaluation

			// 報酬価格を支払う
			l.song_events[i].evaluation_reward[song] += float64(l.evaluation_cost)
			me.energy -= float64(l.evaluation_cost)

			// 記憶に追加
			l.memory = append(l.memory, song)

			// 集計 (II)
			summery.num_evaluation_all++
			summery.num_evaluation_this++

			// 集計 (IV)
			summery.sum_evaluation += evaluation
		}
	}

	// 全曲を聴いたら初期化
	l.incoming_songs = []*Song{}
	l.song_events = []*Event{}
}
