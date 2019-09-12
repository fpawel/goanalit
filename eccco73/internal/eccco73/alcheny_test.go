package eccco73

import (
	"testing"
	"fmt"
)

func TestPieceWiseLinearApproximation(t *testing.T) {
	m := map[float64]float64{1:1,2:4,3:9,4:16,5:25,6:35}

	fmt.Println(PieceWiseLinearApproximation(m, 5.71))
	fmt.Println(PieceWiseLinearApproximation(m, 0.5))
	fmt.Println(PieceWiseLinearApproximation(m, 6.1))

}
