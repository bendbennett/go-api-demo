package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPresenter_Search(t *testing.T) {
	presenter := NewPresenter()

	createdAt, err := time.Parse(
		time.RFC3339,
		"2015-09-15T14:23:12+07:00")
	if err != nil {
		t.Error(err)
	}

	vm := presenter.viewModel(outputData{
		{
			CreatedAt: createdAt,
			ID:        "0a81dec3-3638-4eb4-b04a-83d744f5f3a8",
			FirstName: "john",
			LastName:  "smith",
		},
	})

	assert.Equal(t, "0a81dec3-3638-4eb4-b04a-83d744f5f3a8", vm[0].ID)
	assert.Equal(t, "john", vm[0].FirstName)
	assert.Equal(t, "smith", vm[0].LastName)
	assert.Equal(t, "2015-09-15T14:23:12+07:00", vm[0].CreatedAt)
}
