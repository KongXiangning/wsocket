package wsocket

import (
	"bytes"
	"github.com/gobwas/ws"
	"github.com/prometheus/common/log"
	"net"
)

func BinarySend(buffer bytes.Buffer, conn net.Conn) {
	Send(ws.OpBinary, buffer, conn)
}

func TextSend(text string, conn net.Conn) {
	Send(ws.OpText, *bytes.NewBufferString(text), conn)
}

func PingSend(buffer bytes.Buffer, conn net.Conn) {
	Send(ws.OpPing, buffer, conn)
}

func PongSend(buffer bytes.Buffer, conn net.Conn) {
	Send(ws.OpPong, buffer, conn)
}

func Send(opcode ws.OpCode, buffer bytes.Buffer, conn net.Conn) {
	var (
		length int
		err    error
	)
	header := ws.Header{
		Rsv:    0,
		Masked: false,
		OpCode: opcode,
	}

	for {
		if buffer.Len() > WriteBufferSize {
			header.Fin = false
			header.Length = WriteBufferSize
			length = WriteBufferSize
		} else {
			header.Fin = true
			length = buffer.Len()
			header.Length = int64(length)
		}

		if err = ws.WriteHeader(conn, header); err != nil {
			log.Errorf("write header has error : %s", err)
		}
		if _, err = conn.Write(buffer.Next(length)); err != nil {
			log.Errorf("send has error : %s", err)
		}
		if header.Fin {
			break
		} else {
			header.OpCode = ws.OpContinuation
		}
	}

	/*ERR:
	log.Errorf("send has error : %s",err)*/
}
