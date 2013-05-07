package gnet

import (
  "net"
  "container/list"
)

// bool
type InfiniteBoolChan struct {
  In chan bool
  Out chan bool
  buffer *list.List
  stop chan struct{}
}

func NewInfiniteBoolChan() *InfiniteBoolChan {
  self := &InfiniteBoolChan{
    In: make(chan bool),
    Out: make(chan bool),
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func (self *InfiniteBoolChan) start() {
  for {
    if self.buffer.Len() > 0 {
      elem := self.buffer.Back()
      value := elem.Value.(bool)
      select {
      case self.Out <- value:
        self.buffer.Remove(elem)
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    } else {
      select {
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    }
  }
}

func (self *InfiniteBoolChan) Stop() {
  close(self.stop)
  for {
    select {
    case <-self.In:
      continue
    default:
      return
    }
  }
}

// []byte
type InfiniteByteSliceChan struct {
  In chan []byte
  Out chan []byte
  buffer *list.List
  stop chan struct{}
}

func NewInfiniteByteSliceChan() *InfiniteByteSliceChan {
  self := &InfiniteByteSliceChan{
    In: make(chan []byte),
    Out: make(chan []byte),
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func (self *InfiniteByteSliceChan) start() {
  for {
    if self.buffer.Len() > 0 {
      elem := self.buffer.Back()
      value := elem.Value.([]byte)
      select {
      case self.Out <- value:
        self.buffer.Remove(elem)
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    } else {
      select {
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    }
  }
}

func (self *InfiniteByteSliceChan) Stop() {
  close(self.stop)
  for {
    select {
    case <-self.In:
      continue
    default:
      return
    }
  }
}

// *net.TCPConn
type InfiniteTCPConnChan struct {
  In chan *net.TCPConn
  Out chan *net.TCPConn
  buffer *list.List
  stop chan struct{}
}

func NewInfiniteTCPConnChan() *InfiniteTCPConnChan {
  self := &InfiniteTCPConnChan{
    In: make(chan *net.TCPConn),
    Out: make(chan *net.TCPConn),
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func (self *InfiniteTCPConnChan) start() {
  for {
    if self.buffer.Len() > 0 {
      elem := self.buffer.Back()
      value := elem.Value.(*net.TCPConn)
      select {
      case self.Out <- value:
        self.buffer.Remove(elem)
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    } else {
      select {
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    }
  }
}

func (self *InfiniteTCPConnChan) Stop() {
  close(self.stop)
  for {
    select {
    case <-self.In:
      continue
    default:
      return
    }
  }
}

// *Session
type InfiniteSessionChan struct {
  In chan *Session
  Out chan *Session
  buffer *list.List
  stop chan struct{}
}

func NewInfiniteSessionChan() *InfiniteSessionChan {
  self := &InfiniteSessionChan{
    In: make(chan *Session),
    Out: make(chan *Session),
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func NewInfiniteSessionChanWithOutChan(out chan *Session) *InfiniteSessionChan {
  self := &InfiniteSessionChan{
    In: make(chan *Session),
    Out: out,
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func (self *InfiniteSessionChan) start() {
  for {
    if self.buffer.Len() > 0 {
      elem := self.buffer.Back()
      value := elem.Value.(*Session)
      select {
      case self.Out <- value:
        self.buffer.Remove(elem)
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    } else {
      select {
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    }
  }
}

func (self *InfiniteSessionChan) Stop() {
  close(self.stop)
  for {
    select {
    case <-self.In:
      continue
    default:
      return
    }
  }
}

// *ConnPool
type InfiniteConnPoolChan struct {
  In chan *ConnPool
  Out chan *ConnPool
  buffer *list.List
  stop chan struct{}
}

func NewInfiniteConnPoolChan() *InfiniteConnPoolChan {
  self := &InfiniteConnPoolChan{
    In: make(chan *ConnPool),
    Out: make(chan *ConnPool),
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func (self *InfiniteConnPoolChan) start() {
  for {
    if self.buffer.Len() > 0 {
      elem := self.buffer.Back()
      value := elem.Value.(*ConnPool)
      select {
      case self.Out <- value:
        self.buffer.Remove(elem)
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    } else {
      select {
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    }
  }
}

func (self *InfiniteConnPoolChan) Stop() {
  close(self.stop)
  for {
    select {
    case <-self.In:
      continue
    default:
      return
    }
  }
}

// *Conn
type InfiniteConnChan struct {
  In chan *Conn
  Out chan *Conn
  buffer *list.List
  stop chan struct{}
}

func NewInfiniteConnChan() *InfiniteConnChan {
  self := &InfiniteConnChan{
    In: make(chan *Conn),
    Out: make(chan *Conn),
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func (self *InfiniteConnChan) start() {
  for {
    if self.buffer.Len() > 0 {
      elem := self.buffer.Back()
      value := elem.Value.(*Conn)
      select {
      case self.Out <- value:
        self.buffer.Remove(elem)
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    } else {
      select {
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    }
  }
}

func (self *InfiniteConnChan) Stop() {
  close(self.stop)
  for {
    select {
    case <-self.In:
      continue
    default:
      return
    }
  }
}

// ToSend
type InfiniteToSendChan struct {
  In chan ToSend
  Out chan ToSend
  buffer *list.List
  stop chan struct{}
}

func NewInfiniteToSendChan() *InfiniteToSendChan {
  self := &InfiniteToSendChan{
    In: make(chan ToSend),
    Out: make(chan ToSend),
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func (self *InfiniteToSendChan) start() {
  for {
    if self.buffer.Len() > 0 {
      elem := self.buffer.Back()
      value := elem.Value.(ToSend)
      select {
      case self.Out <- value:
        self.buffer.Remove(elem)
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    } else {
      select {
      case value := <-self.In:
        self.buffer.PushBack(value)
      case <-self.stop:
        return
      }
    }
  }
}

func (self *InfiniteToSendChan) Stop() {
  close(self.stop)
  for {
    select {
    case <-self.In:
      continue
    default:
      return
    }
  }
}

// uint64
type InfiniteUint64Chan struct {
  In chan uint64
  Out chan uint64
  buffer *list.List
  stop chan struct{}
}

func NewInfiniteUint64Chan() *InfiniteUint64Chan {
  self := &InfiniteUint64Chan{
    In: make(chan uint64),
    Out: make(chan uint64),
    buffer: list.New(),
    stop: make(chan struct{}),
  }
  go self.start()
  return self
}

func (self *InfiniteUint64Chan) start() {
  for {
    if self.buffer.Len() > 0 {
      elem := self.buffer.Back()
      value := elem.Value.(uint64)
      select {
      case self.Out <- value:
        self.buffer.Remove(elem)
      case value := <-self.In:
        self.buffer.PushFront(value)
      case <-self.stop:
        return
      }
    } else {
      select {
      case value := <-self.In:
        self.buffer.PushBack(value)
      case <-self.stop:
        return
      }
    }
  }
}

func (self *InfiniteUint64Chan) Stop() {
  close(self.stop)
  for {
    select {
    case <-self.In:
      continue
    default:
      return
    }
  }
}

