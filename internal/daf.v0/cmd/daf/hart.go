package main

import (
	"bytes"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/gohelp"
	"github.com/sirupsen/logrus"
	"strconv"
)

func testHart(p *Product) error {

	// игнор ошибки - возможно, HART протокол был включен ранее
	_ = dafSendCmdToPlace(p.Place, 0x80, 12)

	if err := EN6408SetConnectionLine(p.Place, EN6408ConnectHart); err != nil {
		return err
	}

	err := func() error {
		hartID, err := hartInit()
		if err != nil {
			return merry.Appendf(err, "инициализация HART протокола")
		}
		logrus.Infof("место %d: HART: ID=% X", p.Place+1, hartID)

		b, err := hartReadConcentration(hartID)
		if err != nil {
			return merry.Append(err, "HART: запрос концентрации")
		}
		logrus.Infof("HART: запрос концентрации: место %d: % X", p.Place+1, b)

		if err := hartSwitchOff(hartID); err != nil {
			return err
		}
		return nil
	}()

	if err := EN6408SetConnectionLine(p.Place, EN6408ConnectRS485); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	_, err = dafReadAtPlace(p.Place)
	return err
}

func hartInit() ([]byte, error) {
	b, err := hartGetResponse([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x02, 0x00, 0x00, 0x00, 0x02,
	}, func(b []byte) error {
		// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28
		// 06 00 00 18 00 00 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 BE
		// 06 00 00 18 00 20 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 9E
		// 06 00 00 18 00 00 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 BE
		if len(b) != 29 {
			return comm.Err.WithMessagef("ожидалось 29 байт, получено %d: % X", len(b), b)
		}
		if !bytes.Equal(b[:4], []byte{0x06, 0x00, 0x00, 0x18}) {
			return comm.Err.WithMessagef("ожидалось 06 00 00 18, % X", b[:4])
		}
		if b[6] != 0xFE {
			return comm.Err.WithMessage("b[6] == 0xFE")
		}

		if bytes.Equal(b[23:27], []byte{0x60, 0x93, 0x60, 0x93, 0x01}) {
			return comm.Err.WithMessagef("b[29:27] != 60 93 60 93 01, % X", b[23:27])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return b[15:18], nil
}

func hartGetResponse(req []byte, parse func([]byte) error) ([]byte, error) {
	offset := 0
	log := gohelp.LogPrependSuffixKeys(log, "hart", "")
	response, err := portHart.GetResponse(log, ctxApp, req, func(_, response []byte) (s string, err error) {
		offset, err = parseHart(response, parse)
		s = strconv.Itoa(offset)
		return
	})
	return response[offset:], err
}

func parseHart(response []byte, parse func([]byte) error) (int, error) {

	if len(response) < 5 {
		return 0, comm.Err.WithMessage("длина ответа меньше 5")
	}
	offset := 0
	for i := 2; i < len(response)-1; i++ {
		if response[i] == 0xff && response[i+1] == 0xff && response[i+2] != 0xff {
			offset = i + 2
			break
		}
	}
	if offset == 0 || offset >= len(response) {
		return 0, comm.Err.WithMessage("ответ не соответствует шаблону FF FF XX")
	}
	result := response[offset:]

	if hartCRC(result) != result[len(result)-1] {
		return 0, comm.Err.WithMessage("не совпадает контрольная сумма")
	}
	return offset, parse(result)
}

func hartSwitchOff(hartID []byte) error {
	// 00 01 02 03 04 05 06 07 08 09 10 11 12
	// 82 22 B4 00 00 01 80 04 46 16 00 00 C1
	req := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x82, 0x22, 0xB4,
		hartID[0], hartID[1], hartID[2],
		0x80, 0x04,
		0x46, 0x16, 0x00, 0x00,
		0x00,
	}
	req[5+12] = hartCRC(req[5 : 12+5])

	_, err := hartGetResponse(req, func(b []byte) error {
		// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14
		// 86 22 B4 00 00 01 80 06 00 00 46 16 00 00 C7
		a := []byte{
			0x86, 0x22, 0xB4, hartID[0], hartID[1], hartID[2], 0x80, 0x06,
		}
		if !bytes.Equal(a, b[:8]) {
			return comm.Err.WithMessagef("ожидалось % X, получено % X", a, b[:8])
		}
		return nil
	})
	return err
}

func hartReadConcentration(id []byte) ([]byte, error) {
	// 00 01 02 03 04 05 06 07 08
	// 82 22 B4 00 00 01 01 00 14
	req := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x82, 0x22, 0xB4,
		id[0], id[1], id[2],
		0x01, 0x00,
		0x82,
	}
	req[8+5] = hartCRC(req[5 : 8+5])

	rpat := []byte{0x86, 0x22, 0xB4, id[0], id[1], id[2], 0x01, 0x07}

	b, err := hartGetResponse(req, func(b []byte) error {
		if len(b) < 16 {
			// нужно сделать паузу, возможно плата тормозит
			//time.Sleep(time.Millisecond * 100)
			return comm.Err.WithMessagef("ожидалось 16 байт, получено % X", b)

		}
		// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14 15
		// 86 22 B4 00 00 01 01 07 00 00 A1 00 00 00 00 B6
		if !bytes.Equal(rpat, b[:8]) {
			return comm.Err.WithMessagef("ожидалось % X, получено % X", rpat, b)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return b[11:15], nil
}

func hartCRC(b []byte) byte {
	c := b[0]
	for i := 1; i < len(b)-1; i++ {
		c ^= b[i]
	}
	return c
}
