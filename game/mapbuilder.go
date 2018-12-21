package game

import (
	"math"
	"math/rand"
)

func random(min, max int) int {
	if min == max {
		return min
	}
	return rand.Intn(max-min) + min
}

type Distribution struct {
	cardID int
	count  int
}

func MakeDistribution(cardID int, count int) Distribution {
	return Distribution{cardID, count}
}

type MapBuilder struct {
	distribution []Distribution
}

func makeMapBuilder(distribution []Distribution) MapBuilder {
	return MapBuilder{distribution}
}

func (b *MapBuilder) setDistribution(distribution []Distribution) {
	b.distribution = distribution
}

func (b *MapBuilder) generateMap() gameMap {
	totalCards := 0
	for i := 0; i < len(b.distribution); i++ {
		totalCards += b.distribution[i].count
	}
	mapSize := int(math.Sqrt(float64(totalCards)))
	i := 0
	mapData := make([]int, mapSize*mapSize)
	for len(b.distribution) > 0 {
		index := 0
		if len(b.distribution)-1 > 0 {
			index = random(0, (len(b.distribution)-1)*1000) % (len(b.distribution) - 1)
		}
		cardType := b.distribution[index].cardID
		b.distribution[index].count--
		if b.distribution[index].count == 0 {
			b.distribution = append(b.distribution[:index], b.distribution[index+1:]...)
		}
		mapData[i] = cardType
		i++
	}
	generated := makeMap(mapData, mapSize)
	return generated
}
