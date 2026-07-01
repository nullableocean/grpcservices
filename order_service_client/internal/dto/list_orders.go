package dto

import "fmt"

type ListOrdersParams struct {
	UserUUID      string
	PageSize      int
	NextPageToken string
}

func (p *ListOrdersParams) Validate() error {
	if p.UserUUID == "" {
		return fmt.Errorf("empty user uuid")
	}

	if p.PageSize <= 0 {
		return fmt.Errorf("page size negative or zero")
	}

	return nil
}
