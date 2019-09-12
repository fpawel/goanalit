package worker

type reader struct {
	place int
}

func (x *reader) Perform(w worker) {

	if w.config.ProductsCount() == 0 {
		w.peer.Send("ERROR", "введите адреса приборов")
		return
	}

	if x.place < 0 || x.place >= w.config.ProductsCount() {
		x.place = 0
	} else {
		x.place++
	}
	if w.config.AddrAt(x.place) == 0 {
		return
	}
	w.readVar(x.place, 0)
	return
}
