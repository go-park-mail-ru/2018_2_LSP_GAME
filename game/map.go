package game

func makePosition(onship bool, id int) position {
	return position{onship, id}
}

type gameMap struct {
	size           int
	mapData        []int
	totalGoldCount int
	goldData       []int
}

func makeMap(mapData []int, size int) gameMap {
	result := gameMap{size, mapData, 0, make([]int, size*size)}
	result.countGoldData()
	return result
}

func (m *gameMap) countGoldData() {
	for i := 0; i < m.size*m.size; i++ {
		if m.mapData[i] == GoldCard {
			m.goldData[i] = 1
			m.totalGoldCount++
		}
	}
}

func (m *gameMap) decreaseGold(i int) {
	m.goldData[i]--
}

func (m *gameMap) getGoldOnTitle(i int) int {
	return m.goldData[i]
}

func (m *gameMap) getTotalGoldCount() int {
	return m.totalGoldCount
}

func (m *gameMap) getCardType(i int) int {
	return m.mapData[i]
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (m *gameMap) getMoveableCards(playerID int, pos position) []position {
	res := map[int]bool{}
	shipPos := (m.size - 1) / 2 * (1 + btoi(playerID > 0)*m.size + btoi(playerID == 1)*m.size - btoi(playerID == 2) + btoi(playerID == 3))
	if pos.onship {
		if playerID == 0 || playerID == 1 {
			return []position{makePosition(false, shipPos-1), makePosition(false, shipPos), makePosition(false, shipPos+1)}
		}
		return []position{makePosition(false, shipPos-m.size), makePosition(false, shipPos), makePosition(false, shipPos+m.size)}
	}
	leftEmpty, rightEmpty, upEmpty, downEmpty := true, true, true, true
	if pos.id%m.size == 0 {
		leftEmpty = false
	}
	if (pos.id+1)%m.size == 0 {
		rightEmpty = false
	}
	if pos.id < m.size {
		upEmpty = false
	}
	if pos.id > m.size*m.size-m.size-1 {
		downEmpty = false
	}
	if leftEmpty {
		res[pos.id-1] = true
		if upEmpty {
			res[pos.id-m.size-1] = true
		}
		if downEmpty {
			res[pos.id+m.size-1] = true
		}
	}
	if rightEmpty {
		res[pos.id+1] = true
		if upEmpty {
			res[pos.id-m.size+1] = true
		}
		if downEmpty {
			res[pos.id+m.size+1] = true
		}
	}
	if upEmpty {
		res[pos.id-m.size] = true
		if leftEmpty {
			res[pos.id-m.size-1] = true
		}
		if rightEmpty {
			res[pos.id-m.size+1] = true
		}
	}
	if downEmpty {
		res[pos.id+m.size] = true
		if leftEmpty {
			res[pos.id+m.size-1] = true
		}
		if rightEmpty {
			res[pos.id+m.size+1] = true
		}
	}

	positions := make([]position, 0)
	for k := range res {
		positions = append(positions, position{false, k})
	}

	if pos.id == shipPos {
		positions = append(positions, position{true, 0})
	}

	return positions
}
