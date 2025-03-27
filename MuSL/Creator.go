package MuSL

import (
	"math"
	"math/rand"
)

// 曲のIDを管理するためのカウンター
var nextSongID int = 0

func GetNewSongID() int {
	nextSongID++
	return nextSongID
}

type Song struct {
	id      int // 曲のユニークID
	genre   []float64
	creator *Agent
}

// DeepCopy は Song のディープコピーを作成します
func (s *Song) DeepCopy() *Song {
	// 新しいジャンル配列を作成
	genreCopy := make([]float64, len(s.genre))
	copy(genreCopy, s.genre)

	// 新しい Song 構造体を返す
	return &Song{
		id:      s.id,      // ID はそのまま
		genre:   genreCopy, // コピーしたジャンル情報
		creator: s.creator, // 作成者はポインタなのでそのまま
	}
}

type Creator struct {
	innovation_rate      float64
	memory               []*Song
	creation_probability float64
	creation_cost        Const64
}

func (c *Creator) Create(agents *[]*Agent, me *Agent, summery *Summery) {
	if rand.Float64() < c.creation_probability {
		// 曲を生成
		genre := make([]float64, 2)

		// innovation rate に従ってジャンルを生成
		// memory からランダムに選んで突然変異

		// memory が空の場合はランダムに生成
		if len(c.memory) == 0 {
			for i := 0; i < len(genre); i++ {
				genre[i] = rand.Float64()
			}
		} else {
			random_index := rand.Intn(len(c.memory))
			for i := 0; i < len(genre); i++ {
				// 革新性を両方向に適用（現在は常に正の方向）
				// 修正：-0.5〜0.5の範囲で変動させ、innovation_rateでスケール
				genre[i] = c.memory[random_index].genre[i] + (rand.Float64()-0.5)*2.0*c.innovation_rate

				// 0 以上 1 未満に収める
				genre[i] = math.Max(0.0, math.Min(1.0, genre[i]))
			}
		}

		// 曲を生成して memory に追加
		song := &Song{
			id:      GetNewSongID(),
			genre:   make([]float64, len(genre)), // ジャンル情報用に新しいスライスを作成
			creator: me,
		}
		// ジャンル情報をディープコピー
		copy(song.genre, genre)
		c.memory = append(c.memory, song)

		// エネルギーを消費
		me.energy -= float64(c.creation_cost)

		// 集計 (I)
		summery.num_song_all++
		summery.num_song_this++
	}
}
