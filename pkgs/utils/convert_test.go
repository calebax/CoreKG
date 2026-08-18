package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertBytes(t *testing.T) {
	res, err := ConvertBytes(1203, UnitMB)
	assert.Nil(t, err)
	t.Log("res: ", res)
}
