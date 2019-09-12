package main

import (
	mdbs "github.com/fpawel/guartutils/modbus"
	"github.com/tarm/serial"
	"gopkg.in/natefinch/npipe.v2"
	"log"
	"os"
	"os/exec"
	"time"

	"bytes"
	"fmt"

	"encoding/binary"
	"errors"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/guartutils/ioutils"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"net"
	"strconv"
)

type App struct {
	pipeConn     net.Conn
	stendAddr    byte
	modbus, hart comport.Port
}

func (x App) NewPipeLogger(addr byte, level int, prefix string, flag int) *log.Logger {
	return log.New(ClientPipeWriter{x.pipeConn, level, addr}, prefix, flag)
}

const usage = "usage: dakhart1.exe [MODBUS] [HART] [ADDR STEND] [ADDR 1]...[ADDR N]"

func main() {
	//testRunClientPipeApp()
	log.SetFlags(log.Lshortfile)

	pipeListener, err := npipe.Listen(`\\.\pipe\$TestHart$`)
	if err != nil {
		log.Fatal(err)
	}
	defer pipeListener.Close()

	var app App

	log.Println("pipeRunner: ожидается")
	app.pipeConn, err = pipeListener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer app.pipeConn.Close()

	log.Println("pipeRunner: соединение установлено")

	logErr := app.NewPipeLogger(0, 0, "", log.Lshortfile)
	log.Print(os.Args)

	if len(os.Args) < 5 {
		logErr.Fatalf("должно быть не менее пяти аргументов: %v, usage: %v", os.Args, usage)
	}

	if app.stendAddr, err = stringToByte(os.Args[3]); err != nil {
		logErr.Fatalf("bad stend addr argument %s: %v, args: %v, usage: %v", os.Args[3], err, os.Args, usage)
	}

	var addrs []byte

	for i, s := range os.Args[4:] {
		v, err := stringToByte(s)
		if err != nil {
			logErr.Fatalf("bad addres argument %d: %v: %v", i, s, v)
		}
		addrs = append(addrs, byte(v))
	}

	app.modbus, err = comport.OpenPort(
		comport.Config{
			Serial: serial.Config{
				ReadTimeout: time.Millisecond,
				Baud:        9600,
				Name:        os.Args[1],
			},
			Fetch: fetch.Config{
				ReadTimeout:     2 * time.Second,
				ReadByteTimeout: time.Millisecond * 50,
				MaxAttemptsRead: 2,
			},
		}, nil)
	if err != nil {
		logErr.Fatal(err)
	}
	defer app.modbus.Close()

	app.hart, err = comport.OpenPort(
		comport.Config{
			Serial: serial.Config{
				Baud:        1200,
				ReadTimeout: time.Microsecond,
				Parity:      serial.ParityOdd,
				StopBits:    serial.Stop1,
				Name:        os.Args[2],
			},
			Fetch: fetch.Config{
				ReadTimeout:     2 * time.Second,
				ReadByteTimeout: time.Millisecond * 100,
				MaxAttemptsRead: 3,
			},
		}, nil)
	if err != nil {
		logErr.Fatal(err)
	}
	defer app.hart.Close()

	var oks, errs []byte

	for _, addr := range addrs {
		err = processAddr(addr, app)
		if err == nil {
			app.NewPipeLogger(addr, 1, "", 0).Print("OK")
			oks = append(oks, addr)
		} else {
			app.NewPipeLogger(addr, 0, "", 0).Print(err)
			errs = append(errs, addr)
		}
	}

	if len(errs) > 0 {
		logger := app.NewPipeLogger(0, 0, "", 0)
		logger.Print("Приборы, не прошедщие проверку HART протокола: ")
		for _, addr := range errs[:len(errs)-1] {
			logger.Printf("%d, ", addr)
		}
		logger.Printf("%d, ", errs[len(errs)-1])

	}

	if len(oks) > 0 {
		logger := app.NewPipeLogger(0, 1, "", 0)
		logger.Print("Приборы, успешно прошедщие проверку HART протокола: ")
		for _, addr := range oks[:len(oks)-1] {
			logger.Printf("%d, ", addr)
		}
		logger.Printf("%d", oks[len(oks)-1])
	}

}

func processAddr(addr byte, app App) error {
	log.Printf("адресс %d", addr)
	logger := log.New(ClientPipeWriter{app.pipeConn, 1, addr}, "", 0)
	if app.stendAddr != 0 {
		logger.Print("переключение канала стенда")
		req := mdbs.Request{Addr: app.stendAddr}.PrepareReadBCD((uint16(addr) - 1) * 2).Bytes()
		_, err := app.modbus.Fetch(req)
		if err != nil {
			return fmt.Errorf("не удалось переключить стенд: %v", err)
		}
	} else {
		logger.Print("переключение канала стенда: пропуск операции, адресс стенда 0?")
	}

	log.Print("включение HART протокола")
	_, err := app.modbus.Fetch(mdbs.Request{Addr: addr}.PrepareWriteCmdBCD(0x80, 1000).Bytes())
	if err == nil {
		logger.Print("включение HART протокола: ОК")
	} else {
		logger.Printf("включение HART протокола: %v. Включен ранее?", err)
	}
	log.Print("инициализация HART")
	id, err := initHart(app.hart)
	if err != nil {
		return fmt.Errorf("инициализация HART: %v", err)
	}
	logger.Print("инициализация HART: ОК")

	log.Print("запрос концентрации HART")
	err = readHartConc(app.hart, id)
	if err != nil {
		return fmt.Errorf("запрос концентрации HART: %v", err)
	}
	logger.Print("запрос концентрации HART: ОК")

	log.Print("выключение HART протокола")
	err = switchHartOff(app.hart, id)
	if err != nil {
		return fmt.Errorf("выключение HART протокола: %v", err)
	}
	logger.Print("выключение: ОК")

	log.Print("запрос концентрации MODBUS")
	_, err = app.modbus.Fetch(mdbs.Request{Addr: addr}.PrepareReadBCD(0).Bytes())
	if err != nil {
		return fmt.Errorf("запрос концентрации MODBUS: %v", err)
	}
	logger.Print("запрос концентрации MODBUS: ОК")
	return nil
}

func stringToByte(s string) (byte, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if v < 0 || v > 255 {
		return 0, errors.New("value must be 0..255")
	}
	return byte(v), nil
}

func testRunClientPipeApp() {
	cmd := exec.Command(os.Getenv("GOPATH") + "/src/fpawel/hart6015/test_client/Project1.exe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func hartCRC(b []byte) byte {
	c := b[0]
	for i := 1; i < len(b)-1; i++ {
		c ^= b[i]
	}
	return c
}

func parseHartResponse(b []byte) (r []byte, err error) {
	if len(b) < 5 {
		err = fmt.Errorf("длина ответа меньше 5")
		return
	}
	ok := false
	for i := 2; i < len(b)-1; i++ {
		if b[i] == 0xff && b[i+1] == 0xff && b[i+2] != 0xff {
			r = b[i+2:]
			ok = true
			break
		}
	}
	if !ok || len(r) == 0 {
		err = errors.New("ответ не соответствует шаблону FF FF XX ...")
	}
	if hartCRC(r) != r[len(r)-1] {
		err = fmt.Errorf("ошибка контрольной суммы")
	}
	return
}

func getHartResponse(hart comport.Port, req []byte) (r []byte, err error) {
	b, err := hart.Fetch(req)
	if err != nil {
		return
	}
	r, err = parseHartResponse(b)
	return
}

func initHart(hart comport.Port) (id []byte, err error) {
	b, err := getHartResponse(hart, []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x02, 0x00, 0x00, 0x00, 0x02,
	})
	if err != nil {
		return
	}

	// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28
	// 06 00 00 18 00 00 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 BE
	// 06 00 00 18 00 20 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 9E
	// 06 00 00 18 00 00 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 BE
	if len(b) != 29 {
		err = fmt.Errorf("ожидалось 29 байт, получено %d", len(b))
		return
	}
	if !bytes.Equal(b[:4], []byte{0x06, 0x00, 0x00, 0x18}) {
		err = fmt.Errorf("ожидалось 06 00 00 18, % X", b[:4])
		return
	}
	if b[6] != 0xFE {
		err = fmt.Errorf("b[6] == 0xFE")
		return
	}

	if bytes.Equal(b[23:27], []byte{0x60, 0x93, 0x60, 0x93, 0x01}) {
		err = fmt.Errorf("b[29:27] != 60 93 60 93 01, % X", b[23:27])
		return
	}
	id = b[15:18]
	return
}

func readHartConc(hart comport.Port, id []byte) error {
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

	for i := 0; i < 10; i++ {
		b, err := getHartResponse(hart, req)
		if err != nil {
			return err
		}
		if len(b) < 16 {
			// нужно сделать паузу, возможно плата тормозит
			// time.Sleep(time.Millisecond * 100)
			return fmt.Errorf("ожидалось 16 байт")

		}
		// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14 15
		// 86 22 B4 00 00 01 01 07 00 00 A1 00 00 00 00 B6
		if !bytes.Equal(rpat, b[:8]) {
			return fmt.Errorf("ожидалось % X", rpat)

		}
		log.Printf("№ %d конц. % X", i+1, b[11:15])
		time.Sleep(time.Millisecond * 200)
	}

	return nil
}

func switchHartOff(hart comport.Port, id []byte) (err error) {
	// 00 01 02 03 04 05 06 07 08 09 10 11 12
	// 82 22 B4 00 00 01 80 04 46 16 00 00 C1
	req := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x82, 0x22, 0xB4,
		id[0], id[1], id[2],
		0x80, 0x04,
		0x46, 0x16, 0x00, 0x00,
		0x00,
	}
	req[5+12] = hartCRC(req[5 : 12+5])

	b, err := getHartResponse(hart, req)
	if err != nil {
		return
	}
	// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14
	// 86 22 B4 00 00 01 80 06 00 00 46 16 00 00 C7
	rpat := []byte{
		0x86, 0x22, 0xB4, id[0], id[1], id[2], 0x80, 0x06,
	}
	if !bytes.Equal(rpat, b[:8]) {
		err = fmt.Errorf("ожидалось % X: % X", rpat, b[:8])
	}

	return
}

type ClientPipeWriter struct {
	net.Conn
	level int
	addr  byte
}

func (x ClientPipeWriter) Write(b []byte) (n int, err error) {
	defer func() {
		if err != nil {
			n, err = os.Stdout.Write(b)
		}
	}()

	// отправка - адрес
	if n, err = x.Conn.Write([]byte{x.addr}); err != nil {
		return
	}

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(x.level))

	// отправка - уровень сообщения
	if n, err = x.Conn.Write(bs); err != nil {
		return
	}

	bt, err := UTF8ToWindows1251(b)
	if err != nil {
		return
	}
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(bt)))
	// отправка - длина сообщения

	if _, err = x.Conn.Write(bs); err != nil {
		return
	}
	// отправка - содержимое сообщения

	if _, err = x.Conn.Write(bt); err != nil {
		return
	}
	return
}

func UTF8ToWindows1251(b []byte) (r []byte, err error) {
	buf := new(bytes.Buffer)
	wToWin1251 := transform.NewWriter(buf, charmap.Windows1251.NewEncoder())
	_, err = io.Copy(wToWin1251, bytes.NewReader(b))
	if err == nil {
		r = buf.Bytes()
	}
	return
}
