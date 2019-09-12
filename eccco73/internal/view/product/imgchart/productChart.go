package imgchart

import (
	"github.com/wcharczuk/go-chart"

	"image"
	"fmt"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/fpawel/eccco73/internal/eccco73"
)

func New(x eccco73.Product, width,height int) (image.Image,error) {
	flash := x.Flash()

	seriesFon := chart.ContinuousSeries{
		Style: chart.Style{
			Show:             true,
		},
	}
	seriesFon.XValues, seriesFon.YValues = flash.SeriesFon()

	seriesSens := chart.ContinuousSeries{
		YAxis:   chart.YAxisSecondary,
		Style: chart.Style{
			Show:             true,
			//StrokeWidth:      chart.Disabled,
		},
	}
	seriesSens.XValues, seriesSens.YValues = flash.SeriesSens()

	gridStyle := chart.Style{
		Show:        true,
		StrokeColor: chart.ColorAlternateGray,
		StrokeWidth: 0.3,
	}


	graphChart := chart.Chart{
		Title: fmt.Sprintf("%s, %s, №%v, %.3f", flash.Date().Format("02.01.2006 15:04"), flash.ProductType(),
			flash.Serial(), flash.Sensitivity()),
		TitleStyle: chart.Style{
			Show:      true,
			FontColor: drawing.ColorFromHex("000080"),
			FontSize:  14,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Left: 10,
			},
		},
		Width:width,
		Height:height,
		XAxis: chart.XAxis{
			Name:      "Temperature,⁰C",
			NameStyle: chart.StyleShow(),
			Style: chart.Style{
				Show: true, //enables / displays the x-axis
			},
			TickPosition: chart.TickPositionBetweenTicks,
			Range: &chart.ContinuousRange{ Min: -130, Max: 130, },
			GridMajorStyle: gridStyle,
			GridMinorStyle: gridStyle,
		},
		YAxis: chart.YAxis{
			Name: "Background current, mkA",
			NameStyle: chart.Style{
				FontColor:chart.ColorBlue,
				Show: true,
			},
			Style: chart.Style{
				Show: true, //enables / displays the x-axis
			},
			Range: seriesYRange1(seriesFon),
			GridMajorStyle: gridStyle,
			GridMinorStyle: gridStyle,
		},
		YAxisSecondary: chart.YAxis{
			Name: "Coefficient of sensitivity, %",
			NameStyle: chart.Style{
				FontColor:chart.ColorGreen,
				Show: true,
			},
			Style: chart.Style{
				Show: true, //enables / displays the secondary y-axis
			},
			Range: seriesYRange1(seriesSens),
			GridMajorStyle: gridStyle,
			GridMinorStyle: gridStyle,
		},
		Series: []chart.Series{
			seriesFon,
			seriesSens,
		},
	}
	collector := &chart.ImageWriter{}

	if err := graphChart.Render(chart.PNG, collector); err != nil {
		return nil,err
	}
	return collector.Image()
}

func minMaxFloat64(xs []float64) (min, max  float64) {
	if len(xs) == 0 {
		return
	}

	max, min = xs[0], xs[0]
	for _, value := range xs {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

func seriesYRange1(x chart.ContinuousSeries) *chart.ContinuousRange{
	min, max := minMaxFloat64(x.YValues)
	if min == max {
		return &chart.ContinuousRange{Min:0, Max:1 }
	}
	d := 0.1 * (max - min)
	return &chart.ContinuousRange{Min:min - d, Max:max + d }
}


