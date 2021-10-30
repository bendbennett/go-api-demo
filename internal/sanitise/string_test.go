package sanitise

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_AlphaWithHyphen(t *testing.T) {
	input := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~!@#$%^&*()-_+={}[]\\|<,>.?/\"';:`"
	expectedOutput := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"

	output, err := AlphaWithHyphen(input)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}
