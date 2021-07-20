package rand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRand_GenerateShortID(t *testing.T) {
	id := GenerateShortID()
	assert.True(t, len(id) >= 21)
}

func TestRand_GenerateRandomString(t *testing.T) {
	random, err := GenerateRandomString(10)
	assert.Nil(t, err)
	assert.Len(t, random, 16)
}
