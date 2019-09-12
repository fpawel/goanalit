package kgsdum

import "fmt"

type Gas int

type gasTestPt struct {
	gas Gas
	n   int
}

const (
	Gas1 Gas = 1 + iota
	Gas2
	Gas3
	Gas4
)

func (x Gas) String() string {
	return fmt.Sprintf("ПГС%d", x)
}

func (x gasTestPt) String() string {
	return fmt.Sprintf("№%d %v", x.n, x.gas)
}

var testGases = []Gas{
	Gas1, Gas2, Gas3, Gas4, Gas1,
}

func TestGases() []Gas {
	return []Gas{
		Gas1, Gas2, Gas3, Gas4, Gas1,
	}
}
