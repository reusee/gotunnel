package gnet

import (
  "net"
  "encoding/binary"
  "bytes"
  "time"
  "math/rand"
)

type Conn struct {
  id uint64
  conn *net.TCPConn
  pool *ConnPool
  in chan []byte
  stop chan struct{}

  bytesReadCollect *InfiniteUint64Chan
  bytesSentCollect *InfiniteUint64Chan
}

func newConn(conn *net.TCPConn, connPool *ConnPool) *Conn {
  self := &Conn{
    id: uint64(rand.Int63()),
    conn: conn,
    pool: connPool,
    in: make(chan []byte),
    stop: make(chan struct{}),
  }

  go self.startReadChan()
  go self.start()

  return self
}

func (self *Conn) startReadChan() {
  defer close(self.in)
  for {
    data, err := readFrame(self.conn)
    if err != nil {
      return
    }
    self.in <- data
    if self.bytesReadCollect != nil {
      self.bytesReadCollect.In <- uint64(len(data))
    }
  }
}

func (self *Conn) start() {
  self.log("start")

  heartBeat := time.Tick(time.Second * 2)
  tick := 0
  var err error
  LOOP:
  for {
    select {
    case packet, ok := <-self.in: // read from conn
      if !ok {
        break LOOP
      }
      self.handlePacket(packet)
    case toSend := <-self.pool.sendQueue.Out:
      err = self.handleSend(toSend)
      if err != nil {
        break LOOP
      }
    case data := <-self.pool.rawSendQueue.Out:
      err = self.handleRawSend(data)
      if err != nil {
        break LOOP
      }
    case <-heartBeat:
      err = self.ping()
      if err != nil {
        break LOOP
      }
      //self.log("tick %d", tick)
    case <-self.stop:
      return
    }
    tick++
  }

  self.log("stop")
  if self.pool.deadConnNotify != nil {
    self.pool.deadConnNotify.In <- true
  }
  self.pool.deadConnChan.In <- self
}

func (self *Conn) handlePacket(packet []byte) {
  packetType := packet[0]
  switch packetType {
  case PACKET_TYPE_SESSION:
    self.handleSessionPacket(packet[1:])
  case PACKET_TYPE_INFO:
    self.handleInfoPacket(packet[1:])
  case PACKET_TYPE_PING:
  default:
    self.log("unknown packet type %d\n", packetType)
  }
}

func (self *Conn) handleSessionPacket(packet []byte) {
  payload, sessionId := parseSessionPacket(packet, self.pool.byteKeys, self.pool.uint64Keys)
  session := self.pool.sessions[sessionId]
  if session == nil {
    session = self.pool.newSession(sessionId)
    if session == nil { // session already closed
      return
    }
  }
  if session.closed {
    return
  }
  session.incomingChan.In <- payload
}

func (self *Conn) handleInfoPacket(packet []byte) {
  var sessionId uint64
  //
  entryLen := 37
  entryNum := int(len(packet) / entryLen)
  for i := 0; i < entryNum; i++ {
    binary.Read(bytes.NewReader(packet[i * entryLen : i * entryLen + 8]), binary.BigEndian, &sessionId)
    session := self.pool.sessions[sessionId]
    if session == nil || session.closed {
      continue
    }
    payload := packet[i * entryLen + 8 : (i + 1) * entryLen]
    session.incomingChan.In <- payload
  }
}

func (self *Conn) handleSend(toSend ToSend) error {
  if toSend.tag == DATA && toSend.session.remoteReadState == ABORT {
    return nil
  }
  payload := assembleSessionPacket(toSend.session.id, toSend.data, self.pool.byteKeys, self.pool.uint64Keys)
  err := writeFrame(self.conn, payload)
  if err != nil {
    self.pool.sendQueue.In <- toSend
    return err
  }
  if self.bytesSentCollect != nil {
    self.bytesSentCollect.In <- uint64(len(payload))
  }
  return nil
}

func (self *Conn) handleRawSend(data []byte) error {
  err := writeFrame(self.conn, data)
  if err != nil {
    self.pool.rawSendQueue.In <- data
    return err
  }
  if self.bytesSentCollect != nil {
    self.bytesSentCollect.In <- uint64(len(data))
  }
  return nil
}

func (self *Conn) ping() (err error) {
  err = writeFrame(self.conn, []byte{PACKET_TYPE_PING})
  if err != nil {
    if self.bytesSentCollect != nil {
      self.bytesSentCollect.In <- uint64(1)
    }
  }
  return
}

func (self *Conn) Stop() {
  self.conn.Close()
  close(self.stop)
}

func (self *Conn) log(f string, vars ...interface{}) {
  if DEBUG {
    colorp("33", ps("CONN %d ", self.id) + f, vars...)
  }
}
