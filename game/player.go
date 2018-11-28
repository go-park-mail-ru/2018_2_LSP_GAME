package game

type player struct {
	score   uint
	pirates []pirate
}

func (p *player) incScore() {
	p.score++
}

func (p *player) getScore() uint {
	return p.score
}

func (p *player) addPirate(card position) {
	p.pirates = append(p.pirates, makePirate(card))
}

func (p *player) addPirates(n int, card position) {
	for i := 0; i < n; i++ {
		p.pirates = append(p.pirates, makePirate(card))
	}
}

func (p *player) movePirate(i int, card position) {
	p.pirates[i].setCard(card)
}

func (p *player) getPirates() []pirate {
	return p.pirates
}

func (p *player) getPirate(i int) pirate {
	return p.pirates[i]
}
