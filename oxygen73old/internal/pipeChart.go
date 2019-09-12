package internal

import (
	"encoding/binary"
	"fmt"
	"gopkg.in/natefinch/npipe.v2"
	"net"
	"os/exec"
	"time"
)

const (
	PipeChartCmdAddCurrentTimeValue     = 1
	PipeChartCmdSetRadiogroup1Itemindex = 2
	PipeChartCmdRestoreTimeValue        = 3
)

type PipeChartServer struct {
	quit     chan struct{}
	data     chan *chartPipeMessage
	pipeConn net.Conn
}

type chartPipeMessage struct {
	cmd  byte
	data []byte
}

func NewPipeChartServer() (x *PipeChartServer) {
	x = &PipeChartServer{
		quit: make(chan struct{}),
		data: make(chan *chartPipeMessage),
	}
	return
}

func (x *PipeChartServer) Stop() {
	x.quit <- struct{}{}

}

func (x *PipeChartServer) SendTimeValue(cmd byte, timeValue TimeValue) {
	x.data <- &chartPipeMessage{
		cmd,
		timeValue.serializeToSendToChartPipe(),
	}
}

func (x *PipeChartServer) Send(cmd byte, b []byte) {
	x.data <- &chartPipeMessage{cmd, b}
}

func (x *PipeChartServer) Run(handleTerminated func()) {

	pipeListener, err := npipe.Listen(`\\.\pipe\$Oxygen73Chart$`)
	if err != nil {
		logger.Panic(err)
	}

	//cmdClientApp := exec.Command("c:/projects/oxychart73/Win32/Debug/oxychart.exe")
	cmdClientApp := exec.Command("oxychart.exe")
	err = cmdClientApp.Start()
	if err != nil {
		logger.Panicln("oxychart.exe", err)
	}

	terminatedClientApp := make(chan error, 1)
	go func() {
		terminatedClientApp <- cmdClientApp.Wait()
	}()

	x.pipeConn, err = pipeListener.Accept()
	if err != nil {
		logger.Panicln("pipeRunner error:", err)
	}

	for {
		select {
		case <-x.quit:
			return
		case err := <-terminatedClientApp:
			logger.Println("oxychart.exe terminated:", err)
			handleTerminated()
		case msg := <-x.data:
			x.send(msg)
		}
	}
}

func (x *PipeChartServer) send(msg *chartPipeMessage) error {
	// количество байт данных
	dataLen := len(msg.data)
	bsCount := make([]byte, 4)
	binary.LittleEndian.PutUint32(bsCount, uint32(dataLen))

	bss := [][]byte{
		{msg.cmd}, // код команды
		bsCount,
		msg.data,
	}
	// concat bss to bs
	var bs []byte
	for _, xs := range bss {
		bs = append(bs, xs...)
	}

	x.pipeConn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	nWritten, err := x.pipeConn.Write(bs)
	if err != nil {
		return err
	}
	if nWritten != len(bs) {
		return fmt.Errorf("nWritten [%d] != len(bs) [%d]", nWritten, len(bs))
	}
	return nil
}
