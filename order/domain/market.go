package domain

type Market struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func NewMarket(id int64, name string) *Market {
	return &Market{
		Id:   id,
		Name: name,
	}
}
