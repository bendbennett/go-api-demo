package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateStruct(t *testing.T) {
	type InnerStruct struct {
		StringSlice []string `json:"field_string_slice" validate:"required,min=1,dive,max=3"`
		Int         int      `json:"field_int" validate:"required,min=0,max=10"`
	}

	type OuterStruct struct {
		String      string      `json:"field_string" validate:"required,min=5"`
		InnerStruct InnerStruct `json:"field_inner_struct" validate:"required"`
	}

	outerStruct := OuterStruct{
		String: "four",
		InnerStruct: InnerStruct{
			Int:         11,
			StringSlice: []string{"four"},
		},
	}

	validator, err := NewValidator()
	if err != nil {
		t.Error(err)
	}

	expected := map[string]string{
		"field_string":                             "field_string must be at least 5 characters in length",
		"field_inner_struct.field_int":             "field_int must be 10 or less",
		"field_inner_struct.field_string_slice[0]": "field_string_slice[0] must be a maximum of 3 characters in length",
	}

	assert.Equal(t, expected, validator.ValidateStruct(outerStruct))
}
