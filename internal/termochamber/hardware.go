package termochamber

type Hardware interface {
	Start(responseGetter) error
	Stop(responseGetter) error
	Setup(responseGetter, float64) error
	Read(responseGetter) (float64, error)
}

type responseGetter interface {
	GetResponse([]byte) ([]byte, error)
}
