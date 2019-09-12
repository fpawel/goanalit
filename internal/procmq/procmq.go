package procmq

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/natefinch/npipe.v2"
	"sync"
	"sync/atomic"
)

type ProcessMQ struct {
	listener struct {
		masterToPeer, peerToMaster *npipe.PipeListener
	}
	conn struct {
		masterToPeer, peerToMaster Conn
	}
	// флаг, позволяющий проверить в рантайме факт установки соединения с "партнёром"
	// устанавливается после установки соединения и больше не изменияется
	connected bool
	// атомарный флаг, позволяющий проверить в рантайме факт обрыва соединения с "партнёром"
	// устанавливается после разрыва соединения и больше не изменияется
	// обрыв соединения происходит после получения команды "CLOSE"
	closed int32
	// зарегестрированные обработчики команд от "партнёра"
	// регистрация обработчиков выполняется до установки соединения с "партнёром"
	handlers map[string][]Handler
	// канал получения данных для отправки "партнёру"
	chSend             chan sendData
	chSendAndGetAnswer chan sendAndGetAnswer
}

type Handler func([]byte) interface{}

func MustOpen(pipeStr string) *ProcessMQ {
	x, e := Open(pipeStr)
	if e != nil {
		panic(e)
	}
	return x
}

func Open(pipeStr string) (x *ProcessMQ, err error) {
	x = &ProcessMQ{
		handlers:           make(map[string][]Handler),
		chSend:             make(chan sendData),
		chSendAndGetAnswer: make(chan sendAndGetAnswer),
	}
	if x.listener.masterToPeer, err = npipe.Listen(fmt.Sprintf(`\\.\pipe\$%s_MASTER_TO_PEER$`, pipeStr)); err != nil {
		return
	}
	x.listener.peerToMaster, err = npipe.Listen(fmt.Sprintf(`\\.\pipe\$%s_PEER_TO_MASTER$`, pipeStr))
	return
}

func (x *ProcessMQ) Handle(msg string, h Handler) {
	if x.connected {
		panic("can not add handler after connected")
	}
	handlers, _ := x.handlers[msg]
	x.handlers[msg] = append(handlers, h)
}

func (x *ProcessMQ) Send(msg string, v interface{}) {
	if !x.connected {
		panic("must be connected")
	}

	if atomic.LoadInt32(&x.closed) != 0 {
		return // если соединение оборвано, не отправлять в канал chSend
	}
	x.chSend <- sendData{
		v:   v,
		msg: msg,
	}
}

func (x *ProcessMQ) SendAndGetAnswer(msg string, v interface{}) string {
	if !x.connected {
		panic("must be connected")
	}
	ch := make(chan string)

	if atomic.LoadInt32(&x.closed) != 0 {
		return "" // если соединение оборвано, не отправлять в канал chSendAndGetAnswer
	}
	x.chSendAndGetAnswer <- sendAndGetAnswer{
		v:        v,
		msg:      msg,
		chAnswer: ch,
	}
	return <-ch
}

func (x *ProcessMQ) Connect() (err error) {

	if x.connected {
		panic("already connected")
	}

	if Conn, err = x.listener.masterToPeer.Accept(); err != nil {
		return
	}

	Conn, err = x.listener.peerToMaster.Accept()
	if err == nil {
		x.connected = true
	}

	return
}

func (x *ProcessMQ) Run() error {

	if !x.connected {
		panic("not connected")
	}

	chCloseSend := make(chan struct{}) // канал для закрытия цикла сообщений для отправки данных "партнёру"
	wgSend := new(sync.WaitGroup)      // ожидание закрытия цикла сообщений для отправки данных "партнёру"
	var errWrite, errRead error        // ошибка отправки, приёма
	go func() {                        // цикл сообщений отправки данных "партнёру"
		wgSend.Add(1)
		defer wgSend.Done()
		for {
			select {

			case <-chCloseSend:
				return

			case d := <-x.chSend:
				if errWrite = x.conn.masterToPeer.WriteMsgJSON(d.msg, d.v); errWrite != nil {
					return
				}
			case d := <-x.chSendAndGetAnswer:
				if errWrite = x.conn.masterToPeer.WriteMsgJSON(d.msg, d.v); errWrite != nil {
					return
				}
				var s string
				if s, errRead = x.conn.masterToPeer.ReadString(); errRead != nil {
					return
				}
				d.chAnswer <- s
			}
		}
	}()

	x.connected = true

	for {

		msg, b, err := x.conn.peerToMaster.ReadMsgJSON()
		if err != nil {
			errRead = err
			break
		}
		if msg == "CLOSE" {
			atomic.AddInt32(&x.closed, 1)
			break
		}
		handlers, ok := x.handlers[msg]
		if !ok {
			panic(fmt.Sprintf("handler not found for message %q", msg))
		}
		for _, h := range handlers {
			v := h(b)
			switch v := v.(type) {
			case nil:
				continue
			case []byte:
				x.conn.peerToMaster.Write(v)
			case string:
				x.conn.peerToMaster.WriteString(v)
			default:
				x.conn.peerToMaster.WriteJSON(v)
			}
		}
	}
	chCloseSend <- struct{}{}
	wgSend.Wait()
	return x.close(errRead, errWrite)
}

func (x *ProcessMQ) close(err, errWrite error) error {
	if err == nil && errWrite == nil {
		if err = x.conn.masterToPeer.Close(); err != nil {
			return err
		}
		if err = x.conn.peerToMaster.Close(); err != nil {
			return err
		}
		if err = x.listener.masterToPeer.Close(); err != nil {
			return err
		}
		return x.listener.peerToMaster.Close()
	}

	if err != nil && errWrite != nil {
		return errors.New(err.Error() + ", " + errWrite.Error())
	}
	if err == nil {
		return errWrite
	} else {
		return err
	}
}

type sendData struct {
	msg string
	v   interface{}
}

type sendAndGetAnswer struct {
	msg      string
	v        interface{}
	chAnswer chan string
}
