package create

import "time"

type p struct {
}

type presenter interface {
	viewModel(data outputData) viewModel
}

var _ presenter = (*p)(nil)

func NewPresenter() presenter {
	return &p{}
}

type viewModel struct {
	ID        string
	FirstName string
	LastName  string
	CreatedAt string
}

func (p *p) viewModel(od outputData) viewModel {
	return viewModel{
		ID:        od.ID,
		FirstName: od.FirstName,
		LastName:  od.LastName,
		CreatedAt: od.CreatedAt.Format(time.RFC3339),
	}
}
