package main

import (
	"bytes"
	productsView2 "github.com/fpawel/goanalit/internal/kgsdum/productsView"
	"github.com/lxn/walk"
	"image"
	_ "image/png"
	"log"
)

func AssetImage(path string) walk.Image {

	b, err := Asset(path)
	if err != nil {
		log.Fatalln(err, path)
	}

	x, s, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatalln(err, s, path)
	}
	r, err := walk.NewBitmapFromImage(x)
	if err != nil {
		log.Fatalln(err, s, path)
	}
	return r

}

var ImgCheckmarkPng16 = AssetImage("assets/png16/checkmark.png")
var ImgErrorPng16 = AssetImage("assets/png16/error.png") //
//var ImgCloudPng16 = AssetImage("assets/png16/cloud.png")

//var ImgQuestionPng16 = AssetImage("assets/png16/question.png")
//var ImgForwardPng16 = AssetImage("assets/png16/forward.png")

func init() {

	productsView2.ImgCalendarYearPng16 = AssetImage("assets/png16/calendar-year.png")
	productsView2.ImgCalendarMonthPng16 = AssetImage("assets/png16/calendar-month.png")
	productsView2.ImgCalendarDayPng16 = AssetImage("assets/png16/calendar-day.png")
	productsView2.ImgErrorPng16 = ImgErrorPng16
	productsView2.ImgPartyNodePng16 = AssetImage("assets/png16/folder2.png")
	productsView2.ImgCheckmarkPng16 = ImgCheckmarkPng16
	productsView2.ImgProductNodePng16 = AssetImage("assets/png16/folder1.png")
	productsView2.ImgWindowIcon = NewIconFromResourceId(IconDBID)
}
