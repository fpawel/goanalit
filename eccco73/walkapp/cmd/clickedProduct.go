package main

import (
	"github.com/lxn/walk"
	"image"
	"image/color"
	"log"
	"github.com/fpawel/eccco73/internal/view/product/imgchart"
)

func (x *AppMainWindow) invalidateChart() {

	if !x.ivChart.Visible() {
		return
	}
	x.ivChart.SizeChanged().Detach(0)
	width, height := x.ivChart.Width(), x.ivChart.Height()
	if width > 800 {
		width = 800
	}
	if height > 800 {
		height = 800
	}

	if len(x.clickedProduct.FlashBytes) == 0 {
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				img.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
		x.setChartBitmap(img)
	} else {
		img, err := imgchart.New(x.clickedProduct, width, height)
		if err != nil {
			log.Fatal(err)
		}
		x.setChartBitmap(img)
	}
}

func (x *AppMainWindow) setChartBitmap(img image.Image) {
	bitmap, err := walk.NewBitmapFromImage(img)
	if err != nil {
		log.Fatalln(err)
	}
	x.ivChart.SetImage(bitmap)
	x.ivChart.SizeChanged().Attach(x.invalidateChart)
}

func (x *AppMainWindow) ShowClickedProduct() {
	x.ivChart.SetVisible(true)
	x.app.mw.btnCloseClickedProduct.SetVisible(true)
	x.invalidateChart()
}

func (x *AppMainWindow) HideClickedProduct() {
	x.ivChart.SetVisible(false)
	x.app.mw.btnCloseClickedProduct.SetVisible(false)
}
