package epoll

import (
	"net"
	"sync"
)

type HandlerConn func(conn net.Conn)

type Workerpool struct {
	workers   int
	maxTask   int
	taskQueue chan net.Conn

	lock   sync.Mutex
	closed bool

	connFunc HandlerConn
}

func NewPool(w int, t int, handler HandlerConn) *Workerpool {
	return &Workerpool{
		workers:   w,
		maxTask:   t,
		taskQueue: make(chan net.Conn, t),
		connFunc:  handler,
	}
}

func (wp *Workerpool) Start() {
	for i := 0; i < wp.workers; i++ {
		go wp.StartWorker()
	}
}

func (wp *Workerpool) Close() {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.closed = true
	close(wp.taskQueue)
}

func (wp *Workerpool) AddTask(conn net.Conn) {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	if wp.closed {
		return
	}
	wp.taskQueue <- conn
}

func (wp *Workerpool) StartWorker() {
	for {
		select {
		case conn := <-wp.taskQueue:
			wp.connFunc(conn)
		}
	}
}
