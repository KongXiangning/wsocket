package wsocket

import (
	"bytes"
	"github.com/gobwas/ws"
	"io"
	"log"
	"net"
)

const (
	ReadBufferSize  = 4096
	WriteBufferSize = 4096
)

type handlerInterface interface {
	preHandler(buffer *bytes.Buffer)
	binaryHandler(buffer bytes.Buffer, conn net.Conn)
	textHandler(text string, conn net.Conn)
	pongHandler(buffer bytes.Buffer, conn net.Conn)
	pingHandler(buffer bytes.Buffer, conn net.Conn)
}

type WsHandler struct {
	PreHandler    func(buffer *bytes.Buffer)
	BinaryHandler func(buffer bytes.Buffer, conn net.Conn)
	TextHandler   func(text string, conn net.Conn)
	PongHandler   func(buffer bytes.Buffer, conn net.Conn)
	PingHandler   func(buffer bytes.Buffer, conn net.Conn)
}

func (w *Wserver) handler(conn net.Conn) {
	var (
		opcode ws.OpCode
		err    error
		header ws.Header
		buf    []byte
		buffer bytes.Buffer
	)
	defer w.epoller.ConnSet.Remove(conn.RemoteAddr())
	init := true

	for {
		if header, err = ws.ReadHeader(conn); err != nil {
			log.Printf("connection err : %s\n", err)
			w.epoller.Remove(conn)
			conn.Close()
			return
		}
		if init {
			buf = make([]byte, header.Length)
			opcode = header.OpCode
			init = false
		}
		_, err = io.ReadFull(conn, buf[0:header.Length])
		if header.Masked {
			ws.Cipher(buf, header.Mask, 0)
		}
		buffer.Write(buf[0:header.Length])
		if header.Fin {
			break
		}
	}
	w.handlerImpl.encryptHandler(&buffer)
	switch opcode {
	case ws.OpBinary:
		go w.handlerImpl.binaryHandler(buffer, conn)
	}

}

func (h *WsHandler) binaryHandler(buffer bytes.Buffer, conn net.Conn) {
	if h.BinaryHandler != nil {
		h.BinaryHandler(buffer, conn)
	} else {
		BinarySend(buffer, conn)
	}
}

func (h *WsHandler) textHandler(text string, conn net.Conn) {
	if h.TextHandler != nil {
		h.TextHandler(text, conn)
	} else {
		TextSend(text, conn)
	}
}

func (h *WsHandler) pingHandler(buffer bytes.Buffer, conn net.Conn) {
	if h.PingHandler != nil {
		h.PingHandler(buffer, conn)
	} else {
		PingSend(buffer, conn)
	}
}

func (h *WsHandler) pongHandler(buffer bytes.Buffer, conn net.Conn) {
	if h.PongHandler != nil {
		h.PongHandler(buffer, conn)
	} else {
		PongSend(buffer, conn)
	}
}

func (h *WsHandler) preHandler(buffer *bytes.Buffer) {
	if h.PreHandler != nil {
		h.PreHandler(buffer)
	}
}

func (w *Wserver) SetHandlerImpl(handler handlerInterface) {
	w.handlerImpl = handler
}
