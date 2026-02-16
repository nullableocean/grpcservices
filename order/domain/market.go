package domain

type Market struct {
	id   int64
	name string
}

func NewMarket(id int64, name string) *Market {
	return &Market{
		id:   id,
		name: name,
	}
}

func (m *Market) Id() int64 {
	return m.id
}

func (m *Market) Name() string {
	return m.name
}
