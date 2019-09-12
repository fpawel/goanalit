package termochamber

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

func getResponse(reader responseGetter, s string) (float64, error) {
	s = fmt.Sprintf("\x02%s\r\n", s)
	b, err := reader.GetResponse([]byte(s))
	if err != nil {
		return 0, newErrStr("нет связи", s, nil)
	}
	if len(b) < 4 {
		return 0, newErrStr("несоответствие протоколу: длина ответа менее 4", s, b)
	}
	if b[0] != 2 {
		return 0, newErrStr("несоответствие протоколу: первый байт ответа не 2", s, b)
	}

	r := string(b)

	if !strings.HasSuffix(r, "\r\n") {
		return 0, newErrStr("несоответствие протоколу: ответ должен оканчиваться байтами 0D 0A", s, b)
	}

	r = r[1 : len(r)-2]

	if strings.HasPrefix(s, "01WRD") && r != "01WRD,OK" {
		return 0, newErrStr("несоответствие протоколу: ответ на запрос 01WRD должен быть 01WRD,OK", s, b)
	}

	if strings.HasPrefix(s, "01RRD") {
		if !strings.HasPrefix(r, "01RRD,OK") {
			return 0, newErrStr("несоответствие протоколу: не удалось считать температуру: ответ на запрос 01RRD должен начинаться со строки 01RRD,OK", s, b)
		}
		xs := regexTemperature.FindAllStringSubmatch(r, -1)
		if len(xs) == 0 {
			return 0, newErrStr("не правильный формат температуры", s, b)
		}
		if len(xs[1]) == 2 {
			return 0, newErrStr("не правильный формат температуры: ожидался код значения температуры и уставки", s, b)
		}
		n, err := strconv.ParseInt(xs[1][1], 16, 64)
		if err != nil {
			err = errors.Wrapf(err, "не правильный формат температуры: %q", xs[1][1])
			return 0, newErr(err, s, b)
		}

		return float64(n) / 10, nil
	}
	return 0, nil
}

var regexTemperature = regexp.MustCompile(`^01RRD,OK,([0-9a-fA-F]{4}),([0-9a-fA-F]{4})$`)

func newErrStr(err string, strReq string, b []byte) error {
	return newErr(errors.New(err), strReq, b)
}

func newErr(err error, strReq string, b []byte) error {
	return errors.Wrapf(hardwareError, "%v: запрос %q: [% X], ответ %q: [% X]",
		err, strReq, []byte(strReq), string(b), b)
}

var hardwareError = errors.New("ошибка термокамеры")

func IsHardwareError(err error) bool {
	return errors.Cause(err) == hardwareError
}
