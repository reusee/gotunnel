package gnet

import (
  "net"
  "sync"
  "math/rand"
  "time"
  "encoding/binary"
)

func init() {
  rand.Seed(time.Now().UnixNano())
}

type Client struct {
  id uint64
  stop chan struct{}
  connPool *ConnPool
  conns int
  Closed bool

  livingConns int
  raddr *net.TCPAddr

  deadConnNotify *InfiniteBoolChan
  connPoolStopNotify *InfiniteConnPoolChan

  BytesRead uint64
  bytesReadCollect *InfiniteUint64Chan
  BytesSent uint64
  bytesSentCollect *InfiniteUint64Chan
}

func NewClient(addr string, key string, conns int) (*Client, error) {
  raddr, err := net.ResolveTCPAddr("tcp", addr)
  if err != nil {
    return nil, err
  }

  self := &Client{
    id: uint64(rand.Int63()),
    stop: make(chan struct{}),
    connPool: newConnPool(key, nil),
    conns: conns,
    raddr: raddr,

    deadConnNotify: NewInfiniteBoolChan(),
    connPoolStopNotify: NewInfiniteConnPoolChan(),

    bytesReadCollect: NewInfiniteUint64Chan(),
    bytesSentCollect: NewInfiniteUint64Chan(),
  }
  self.connPool.deadConnNotify = self.deadConnNotify
  self.connPool.stopNotify = self.connPoolStopNotify.In
  self.connPool.bytesReadCollect = self.bytesReadCollect
  self.connPool.bytesSentCollect = self.bytesSentCollect

  err = self.connect(conns)
  if err != nil {
    return nil, err
  }

  go self.start()

  return self, nil
}

func (self *Client) start() {
  heartBeat := time.Tick(time.Second * 2)
  tick := 0

  LOOP:
  for {
    select {
    case <-self.deadConnNotify.Out:
      if self.livingConns > 0 {
        self.livingConns--
      }
    case <-heartBeat:
      self.log("tick %d %d conns", tick, self.livingConns)
      self.checkConns()
    case <-self.connPoolStopNotify.Out:
      self.log("lost connection to server")
      break LOOP
    case c := <-self.bytesReadCollect.Out:
      self.BytesRead += c
    case c := <-self.bytesSentCollect.Out:
      self.BytesSent += c
    case <-self.stop:
      break LOOP
    }
    tick++
  }

  // finalizer
  self.Closed = true
  self.connPool.Stop()
  self.deadConnNotify.Stop()
  self.connPoolStopNotify.Stop()
  self.bytesReadCollect.Stop()
  self.bytesSentCollect.Stop()
}

func (self *Client) Stop() {
  close(self.stop)
}

func (self *Client) checkConns() {
  if self.livingConns < self.conns && !self.connPool.closed {
    self.connect(self.conns - self.livingConns)
  }
}

func (self *Client) connect(conns int) (err error) {
  wg := new(sync.WaitGroup)
  wg.Add(conns)
  for i := 0; i < conns; i++ {
    go func() {
      defer wg.Done()
      var conn *net.TCPConn
      conn, err = net.DialTCP("tcp", nil, self.raddr)
      if err != nil {
        return
      }
      binary.Write(conn, binary.BigEndian, self.id)
      self.connPool.newConnChan.In <- conn
      self.livingConns++
    }()
  }
  wg.Wait()
  return
}

func (self *Client) NewSession() *Session {
  return self.connPool.newSession(uint64(rand.Int63()))
}

func (self *Client) log(s string, vars ...interface{}) {
  if DEBUG {
    colorp("31", ps("CLIENT %d ", self.id) + s, vars...)
  }
}
