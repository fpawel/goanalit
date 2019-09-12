package eccco73

import (
	"math"
	"sort"
)

type X2Y map[float64]float64

type DeltaCurrentToCoefficientSensitivityPercentFunc func (curFon,curSens float64) float64

func MainTemperatures() []float64{
	return []float64{ -40, -20, 0, 20, 30, 40, 45, 50,	}
}

func Fon3(y X2Y)  {
	if _,ok := y[20]; !ok {
		return
	}
	if _,ok := y[50]; !ok {
		return
	}
	if _,ok := y[-20]; !ok {
		return
	}
	y[40] = (y[50] - y[20]) * 0.5 + y[20]
	y[-40] = y[-20] - 0.5 * (y[20] -y[-20])
	y[0] = y[20] - 0.5 * (y[20] -y[-20])
	y[30] = (y[40] - y[20]) * 0.5 + y[20]
	y[45] = (y[50] - y[40]) * 0.5 + y[40]

}

func Fon2(y X2Y) {
	if _,ok := y[20]; !ok {
		return
	}
	if _,ok := y[50]; !ok {
		return
	}

	y[40] = (y[50] - y[20]) * 0.5 + y[20]
	y[-40] = 0
	y[-20] = y[20] * 0.2
	y[0] = y[20] * 0.5
	y[30] = (y[40] - y[20]) * 0.5 + y[20]
	y[45] = (y[50] - y[40]) * 0.5 + y[40]
}

func Sens2(y X2Y) {
	if _,ok := y[20]; !ok {
		return
	}
	if _,ok := y[50]; !ok {
		return
	}
	y[40] = (y[50]-y[20])*0.5 + y[20]
	y[-40] = 30
	y[-20] = 58
	y[0] = 82
	y[30] = (y[40]-y[20])*0.5 + y[20]
	y[45] = (y[50]-y[40])*0.5 + y[40]
}

func Sens3(y X2Y) X2Y {
	if _,ok := y[20]; !ok {
		return nil
	}
	if _,ok := y[50]; !ok {
		return nil
	}
	if _,ok := y[-20]; !ok {
		return nil
	}
	y[0] = (y[20] - y[-20]) * 0.5 + y[-20]
	y[40] = y[50]-y[20] * 0.5 + y[20]
	y[45] = y[50]-y[40] * 0.5 + y[40]
	y[30] = y[40]-y[20] * 0.5 + y[20]
	y[-40] = 2 * y[-20] - y[0]
	if y[-20] > 0 {
		y[-40] += 1.2 * (45 - y[-20])/( 0.43429 * math.Log(y[-20]) )
	} else if y[-20] < 0.45 * y[20] {
		return nil
	}
	return y
}

func Fon (m PointsMethod, xy X2Y) {
	switch m {
	case PointsMethod2:
		Fon2(xy)
	case PointsMethod3:
		Fon3(xy)
	default:
		panic(m)
	}
}

func Sens (m PointsMethod, xy X2Y) {
	switch m {
	case PointsMethod2:
		Sens2(xy)
	case PointsMethod3:
		Sens3(xy)
	default:
		panic(m)
	}
}

// кусочно-линейная апроксимация
func PieceWiseLinearApproximation(xy X2Y, x float64) float64{

	n := len(xy)
	if n == 0{
		panic("map must be not empty")
	}

	var xs,ys []float64
	{
		var xys [][2]float64
		for x,y := range xy {
			xys = append(xys, [2]float64{x,y})
		}
		sort.Slice(xys, func(i, j int) bool {
			return xys[i][0] < xys[j][0]
		})
		for _,a := range xys{
			xs = append(xs, a[0])
			ys = append(ys, a[1])
		}
	}
	if x < xs[0] {
		return ys[0]
	}
	for i := 1; i<n; i++ {
		if xs[i-1] <= x && x < xs[i] {
			b := (ys[i]-ys[i-1])/(xs[i]-xs[i-1])
			a := ys[i-1] - b*xs[i-1]
			return a + b *x
		}
	}
	return ys[n-1]

}
