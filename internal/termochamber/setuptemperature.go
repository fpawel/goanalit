package termochamber

import (
	"context"
	"time"
)

func WaitForSetupTemperature(tMin, tMax float64, timeout time.Duration, readT func() (float64, error)) error {
	ch := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	go func() {
		defer cancel() // рекомендовано принудительно отменять контекст
		for {
			select {
			case <-ctx.Done():
				ch <- ctx.Err() // таймаут
				return
			default:
				v, err := readT()
				if err != nil {
					ch <- err // ошибка связи
					return
				}
				if v >= tMin && v <= tMax {
					ch <- nil // термокамера вышла на нужную температуру
					return
				}
			}
		}
	}()
	return <-ch
}
