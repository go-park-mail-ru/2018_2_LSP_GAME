package game

func makePosition(onship bool, id int) position {
	return position{onship, id}
}

type gameMap struct {
	Size            int
	mapData         []int
	totalGoldCount  int
	totalGoldUnused int
	goldData        []int
}

func makeMap(mapData []int, size int) gameMap {
	result := gameMap{
		Size:            size,
		mapData:         mapData,
		totalGoldCount:  0,
		totalGoldUnused: 0,
		goldData:        make([]int, size*size),
	}
	result.countGoldData()
	return result
}

func (m *gameMap) countGoldData() {
	for i := 0; i < m.Size*m.Size; i++ {
		if m.mapData[i] == GoldCard {
			m.goldData[i] = 1
			m.totalGoldCount++
		}
	}
	m.totalGoldUnused = m.totalGoldCount
}

func (m *gameMap) decreaseGold(i int) {
	if m.getGoldOnTitle(i) > 0 {
		m.goldData[i]--
		m.totalGoldUnused--
	}
}

func (m *gameMap) getUnusedGoldCount() int {
	return m.totalGoldUnused
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
	shipPos := (m.Size - 1) / 2 * (1 + btoi(playerID > 0)*m.Size + btoi(playerID == 1)*m.Size - btoi(playerID == 2) + btoi(playerID == 3))
	if pos.onship {
		if playerID == 0 || playerID == 1 {
			return []position{makePosition(false, shipPos-1), makePosition(false, shipPos), makePosition(false, shipPos+1)}
		}
		return []position{makePosition(false, shipPos-m.Size), makePosition(false, shipPos), makePosition(false, shipPos+m.Size)}
	}
	leftEmpty, rightEmpty, upEmpty, downEmpty := true, true, true, true
	if pos.id%m.Size == 0 {
		leftEmpty = false
	}
	if (pos.id+1)%m.Size == 0 {
		rightEmpty = false
	}
	if pos.id < m.Size {
		upEmpty = false
	}
	if pos.id > m.Size*m.Size-m.Size-1 {
		downEmpty = false
	}
	if leftEmpty {
		res[pos.id-1] = true
		if upEmpty {
			res[pos.id-m.Size-1] = true
		}
		if downEmpty {
			res[pos.id+m.Size-1] = true
		}
	}
	if rightEmpty {
		res[pos.id+1] = true
		if upEmpty {
			res[pos.id-m.Size+1] = true
		}
		if downEmpty {
			res[pos.id+m.Size+1] = true
		}
	}
	if upEmpty {
		res[pos.id-m.Size] = true
		if leftEmpty {
			res[pos.id-m.Size-1] = true
		}
		if rightEmpty {
			res[pos.id-m.Size+1] = true
		}
	}
	if downEmpty {
		res[pos.id+m.Size] = true
		if leftEmpty {
			res[pos.id+m.Size-1] = true
		}
		if rightEmpty {
			res[pos.id+m.Size+1] = true
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
