package game

type position struct {
	onship bool
	id     int
}

type pirate struct {
	card position
}

func makePirate(card position) pirate {
	return pirate{card}
}

func (p *pirate) setCard(card position) {
	p.card = card
}

func (p *pirate) getCard() position {
	return p.card
}

func (p *pirate) kill() {
	p.card = position{true, 0}
}
