package utils_test

import (
	"testing"

	"github.com/nielskrijger/goboot/utils"
	"github.com/stretchr/testify/assert"
)

func TestRand_GenerateShortID(t *testing.T) {
	id := utils.GenerateShortID()
	assert.True(t, len(id) >= 21)
}

func TestRand_GenerateRandomString(t *testing.T) {
	random, err := utils.GenerateRandomString(10)
	assert.Nil(t, err)
	assert.Len(t, random, 16)
}
