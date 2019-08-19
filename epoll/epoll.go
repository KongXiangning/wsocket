package epoll

import (
	"golang.org/x/sys/unix"
	"net"
	"reflect"
	"sync"
	"syscall"
	"wsocket/util"
)

type Epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
	ConnSet     *util.Set
}

func NewEpoll() (*Epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	return &Epoll{
		fd:          fd,
		connections: make(map[int]net.Conn),
		lock:        &sync.RWMutex{},
		ConnSet:     util.New(),
	}, nil
}

func (e *Epoll) Add(conn net.Conn) error {
	fd := socketFD(conn)
	if err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.EPOLLHUP, Fd: int32(fd)}); err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn

	return nil
}

func (e *Epoll) Remove(conn net.Conn) error {
	fd := socketFD(conn)
	if err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil); err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()

	delete(e.connections, fd)
	return nil
}

func (e *Epoll) Wait() ([]net.Conn, error) {
	events := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		return nil, err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Fd)]
		if e.ConnSet.Add(conn.RemoteAddr()) {
			connections = append(connections, conn)
		}
	}
	return connections, nil
}

func socketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}
