package wsocket

import (
	"github.com/gobwas/ws"
	"log"
	"net"
	"net/http"
	"runtime"
	"wsocket/epoll"
)

type Wserver struct {
	addr        string
	epoller     *epoll.Epoll
	workerpool  *epoll.Workerpool
	handlerImpl handlerInterface
	Ug          ws.Upgrader
}

func Init(workers int, maxTask int, addr string, handler handlerInterface) *Wserver {
	epoller, err := epoll.NewEpoll()
	if err != nil {
		panic(err)
	}
	w := &Wserver{
		addr:        addr,
		epoller:     epoller,
		Ug:          ws.Upgrader{},
		handlerImpl: handler,
	}
	w.workerpool = epoll.NewPool(workers, maxTask, w.handler)
	return w
}

func (w *Wserver) Start() error {

	ln, err := net.Listen("tcp", w.addr)
	if err != nil {
		return err
	}
	defaultUpgrade(&w.Ug)

	go w.wait()
	w.workerpool.Start()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("connction error : %s\n", err)
			continue
		}
		_, err = w.Ug.Upgrade(conn)
		if err != nil {
			log.Printf("upgrade error: %s \n", err)
		}
		log.Println("connction is success:%v", conn)
		w.epoller.Add(conn)
	}
}

func defaultUpgrade(upgrade *ws.Upgrader) {
	if upgrade.OnBeforeUpgrade == nil {
		upgrade.OnBeforeUpgrade = func() (header ws.HandshakeHeader, err error) {
			log.Println("current is run here")
			return ws.HandshakeHeaderHTTP(http.Header{
				"X-Go-Version": []string{runtime.Version()},
			}), nil
		}
	}
}

func (w *Wserver) wait() {
	for {
		connections, err := w.epoller.Wait()
		if err != nil {
			log.Printf("failed to epoll wait %v \n", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			w.workerpool.AddTask(conn)
		}
	}
}
