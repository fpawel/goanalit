package main

import (
	"flag"
	sph "github.com/fpawel/guartutils/comport"
	mdbs "github.com/fpawel/guartutils/modbus"
	"github.com/fpawel/gutils/utils"
	"github.com/tarm/serial"
	"gopkg.in/natefinch/npipe.v2"
	"log"
	"os"
	"os/exec"
	"time"

	"bytes"
	"fmt"
)

func main() {
	//testRunClientPipeApp()
	log.SetFlags(log.Lshortfile)
	pipeListener, err := npipe.Listen(`\\.\pipe\$TestHart$`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("pipeRunner: ожидается")
	pipeConn, err := pipeListener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("pipeRunner: соединение установлено")
	logErr := log.New(ClientPipeWriter{pipeConn, 0}, "", log.Lshortfile)
	logInfo := log.New(ClientPipeWriter{pipeConn, 1}, "", 0)
	//logDebug := log.New(ClientPipeWriter{pipeConn,2}, "", 0)

	modbus := &sph.Port{
		Config: &sph.Config{
			Config: &serial.Config{
				Baud:        9600,
				ReadTimeout: time.Millisecond,
			},
			ReadTimeout:     2 * time.Second,
			ReadByteTimeout: time.Millisecond * 50,
			MaxAttemptsRead: 3,
		},
		Logger: log.New(ClientPipeWriter{pipeConn, 2}, "MODBUS:", 0),
	}
	hart := &sph.Port{
		Config: &sph.Config{
			Config: &serial.Config{
				Baud:        1200,
				ReadTimeout: time.Microsecond,
				Parity:      serial.ParityOdd,
				StopBits:    serial.Stop1,
			},
			ReadTimeout:     2 * time.Second,
			ReadByteTimeout: time.Millisecond * 100,
			MaxAttemptsRead: 5,
		},
		Logger: log.New(ClientPipeWriter{pipeConn, 2}, "HART:", 0),
	}
	flag.StringVar(&modbus.Config.Name, "modbus", "", "COM port MODBUS")
	flag.StringVar(&hart.Config.Name, "hart", "", "COM port HART")
	flag.Parse()

	log.Println("MODBUS", modbus.Config.Name, "HART", hart.Config.Name)

	modbus.Port, err = serial.OpenPort(modbus.Config.Config)
	if err != nil {
		logErr.Fatal(err)
	}
	hart.Port, err = serial.OpenPort(hart.Config.Config)
	if err != nil {
		logErr.Fatal(err)
	}
	defer func() {
		modbus.Close()
		hart.Close()
	}()

	// установить адрес 1
	bs := mdbs.Request{}.PrepareWriteCmdBCD(7, 1).Bytes()
	err = modbus.WriteAndCheckWritenCount(bs)
	if err != nil {
		logErr.Panic(err)
	}

	log.Println(modbus.Config.Name, utils.FormatBytesHex(bs), "установка адреса 1")
	time.Sleep(time.Second)

	log.Println("включение HART протокола")
	modbus.Config.MaxAttemptsRead = 1
	modbus.GetResponseBytes(mdbs.Request{Addr: 1}.PrepareWriteCmdBCD(0x80, 1000).Bytes())
	modbus.Config.MaxAttemptsRead = 3
	logInfo.Print("включение - ОК")

	log.Println("инициализация HART")
	id, err := initHart(hart)
	if err != nil {
		logErr.Fatal(err)
	}
	logInfo.Print("инициализация - ОК")

	log.Println("запрос концентрации HART")
	err = readHartConc(hart, id)
	if err != nil {
		logErr.Fatal(err)
	}
	logInfo.Print("запрос концентрации HART - ОК")

	log.Println("выключение HART протокола")
	err = switchHartOff(hart, id)
	if err != nil {
		logErr.Fatal(err)
	}
	logInfo.Print("выключение - ОК")

	log.Println("запрос концентрации MODBUS")
	_, err = modbus.GetResponseBytes(mdbs.Request{Addr: 1}.PrepareReadBCD(0).Bytes())
	if err != nil {
		logErr.Fatal(err)
	}
	logInfo.Print("запрос концентрации MODBUS - ОК")
	logInfo.Print("успешно")
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
		err = fmt.Errorf("ответ не соответствует шаблону FF FF XX ...")
	}
	if hartCRC(r[:len(r)]) != r[len(r)-1] {
		err = fmt.Errorf("ошибка контрольной суммы")
	}
	return
}

func getHartResponse(hart *sph.Port, req []byte) (r []byte, err error) {
	b, err := hart.GetResponseBytes(req)
	if err != nil {
		return
	}
	r, err = parseHartResponse(b)
	return
}

func initHart(hart *sph.Port) (id []byte, err error) {
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
		err = fmt.Errorf("ожидалось 06 00 00 18, %s", utils.FormatBytesHex(b[:4]))
		return
	}
	if b[6] != 0xFE {
		err = fmt.Errorf("b[6] == 0xFE")
		return
	}

	if bytes.Equal(b[23:27], []byte{0x60, 0x93, 0x60, 0x93, 0x01}) {
		err = fmt.Errorf("b[29:27] != 60 93 60 93 01, %f", utils.FormatBytesHex(b[23:27]))
		return
	}
	id = b[15:18]
	return
}

func readHartConc(hart *sph.Port, id []byte) error {
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
			//time.Sleep(time.Millisecond * 100)
			return fmt.Errorf("ожидалось 16 байт")

		}
		// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14 15
		// 86 22 B4 00 00 01 01 07 00 00 A1 00 00 00 00 B6
		if !bytes.Equal(rpat, b[:8]) {
			return fmt.Errorf("ожидалось %s", utils.FormatBytesHex(rpat))

		}
		log.Println("№", i+1, "конц.", utils.FormatBytesHex(b[11:15]))
		time.Sleep(time.Millisecond * 200)
	}

	return nil
}

func switchHartOff(hart *sph.Port, id []byte) (err error) {
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
		err = fmt.Errorf("ожидалось %s: %s", utils.FormatBytesHex(rpat), utils.FormatBytesHex(b[:8]))
	}

	return
}
