package search

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

type viewModel []usr

type usr struct {
	ID        string
	FirstName string
	LastName  string
	CreatedAt string
}

func (p *p) viewModel(od outputData) viewModel {
	var vm viewModel

	for _, u := range od {
		vm = append(
			vm,
			usr{
				ID:        u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				CreatedAt: u.CreatedAt.Format(time.RFC3339),
			},
		)
	}

	return vm
}
