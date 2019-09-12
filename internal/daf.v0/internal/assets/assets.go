package assets

import (
	"bytes"
	"github.com/lxn/walk"
	"image"
	_ "image/png"
	"log"
)

var (
	ImgPinOn     = Image("assets/png16/pin_on.png")
	ImgPinOff    = Image("assets/png16/pin_off.png")
	ImgForward   = Image("assets/png16/forward.png")
	ImgError     = Image("assets/png16/error.png")
	ImgCheckMark = Image("assets/png16/checkmark.png")
)

func Image(path string) walk.Image {

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
