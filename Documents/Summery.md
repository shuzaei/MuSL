# Summery

## 概要
`Summery`は、シミュレーションの集計結果をまとめるためのクラスです。

主に以下の値を集計します。

A. ジャンル多様性
B. 人数と内訳
C. 志向の平均
D. エネルギーの総量
E. 作成曲数、試聴述べ曲数、イベントの開催数
F. 役割ごとのエネルギーの増減
G. そのイテレーションで行われた評価の平均

- `num_population` int: 人数 (B)
- `num_creaters` int: 作成者の人数 (B)
- `num_listeners` int: 聴取者の人数 (B)
- `num_organizers` int: 運営者の人数 (B)
- `num_song_all` int: いままで作成された楽曲の総数 (E)
- `num_song_this` int: そのイテレーションで作成された楽曲の総数 (E)
- `num_song_now` int: 現在残っているエージェントの楽曲の総数 (E)
- `num_evaluation_all` int: いままで行われた評価の総数 (E)
- `num_evaluation_this` int: そのイテレーションで行われた評価の総数 (E)
- `num_event_all` int: いままで開催されたイベントの総数 (E)
- `num_event_this` int: そのイテレーションで開催されたイベントの総数 (E)
- `avg_innovation` float: 作成者の新規性の平均 (C)
- `avg_novelty_preference` float: 聴取者の新規性好みの平均 (C)
- `sum_evaluation` float: そのイテレーションで行われた評価の合計 (G)
- `avg_evaluation` float: そのイテレーションで行われた評価の平均 (G)
- `total_energy` float: エネルギーの総量 (D)
- `energy_creators` float: 作成者のエネルギーの総量 (F)
- `energy_listeners` float: 聴取者のエネルギーの総量 (F)
- `energy_organizers` float: 運営者のエネルギーの総量 (F)
- `all_genres` list: すべてのジャンル (A)

以下は、Visualizer で計算される値です。（各イテレーションで計算すると、計算量が多くなるため）
- `num_major` int: メジャー曲の数 (A)
- `num_minor` int: マイナー曲の数 (A)

A については、以下のように計算します。
range_threshold より距離が小さい音楽の個数が、dencity_threshold より小さいかどうかで
各曲のメジャー・マイナーを判定します。
ここで、単純に分布が広いことがこの研究における多様性とは限らないため、いわゆる一般的な多様性の指標は集計しないことにします。

## 複雑すぎるので、省略する要素

- 曲といっしょに別の値を集計するといったことはとりあえずしない。