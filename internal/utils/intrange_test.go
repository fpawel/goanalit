package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIntRanges(t *testing.T) {
	assert := assert.New(t)
	xs := ParseIntRanges("1  - 5 7 9 -13 14 15- 17 ")
	assert.EqualValues(xs, []int{1, 2, 3, 4, 5, 7, 9, 10, 11, 12, 13, 14, 15, 16, 17})
	assert.EqualValues(FormatIntRanges(xs), "1-5 7 9-17")
}
