package gnet

import (
  "bytes"
  "encoding/binary"
  "time"
)

type ConnPool struct {
  clientId uint64
  stopNotify chan *ConnPool

  stop chan struct{}
  newConnChan *InfiniteTCPConnChan
  deadConnNotify *InfiniteBoolChan
  deadConnChan *InfiniteConnChan

  byteKeys []byte
  uint64Keys []uint64

  sendQueue *InfiniteToSendChan
  infoChan *InfiniteToSendChan
  rawSendQueue *InfiniteByteSliceChan

  sessions map[uint64]*Session
  newSessionChan *chan *Session
  sessionStopNotify *InfiniteSessionChan
  newSessionRequest chan newSessionReq
  stoppedSessions map[uint64]bool

  conns map[uint64]*Conn
  maxConnNum int

  closed bool

  bytesReadCollect *InfiniteUint64Chan
  bytesSentCollect *InfiniteUint64Chan
}

type newSessionReq struct {
  ret chan *Session
  sessionId uint64
}

func newConnPool(key string, newSessionChan *chan *Session) *ConnPool {
  self := &ConnPool{
    stop: make(chan struct{}),
    newConnChan: NewInfiniteTCPConnChan(),
    deadConnChan: NewInfiniteConnChan(),

    sendQueue: NewInfiniteToSendChan(),
    infoChan: NewInfiniteToSendChan(),
    rawSendQueue: NewInfiniteByteSliceChan(),

    sessions: make(map[uint64]*Session),
    newSessionChan: newSessionChan,
    sessionStopNotify: NewInfiniteSessionChan(),
    newSessionRequest: make(chan newSessionReq),
    stoppedSessions: make(map[uint64]bool),

    conns: make(map[uint64]*Conn),
  }
  self.byteKeys, self.uint64Keys = calculateKeys(key)

  go self.start()

  return self
}

func (self *ConnPool) start() {
  self.log("start")
  infoBuf := new(bytes.Buffer)
  heartBeat := time.Tick(time.Second * 2)
  tick := 0

  LOOP:
  for {
    select {
    case tcpConn := <-self.newConnChan.Out:
      conn := newConn(tcpConn, self)
      conn.bytesReadCollect = self.bytesReadCollect
      conn.bytesSentCollect = self.bytesSentCollect
      self.conns[conn.id] = conn
      if len(self.conns) > self.maxConnNum {
        self.maxConnNum = len(self.conns)
      }

    case info := <-self.infoChan.Out:
      binary.Write(infoBuf, binary.BigEndian, info.session.id)
      infoBuf.Write(info.data)

    case <-heartBeat:
      // send session infos
      info := infoBuf.Bytes()
      frame := new(bytes.Buffer)
      frame.Write([]byte{PACKET_TYPE_INFO})
      frame.Write(info)
      self.rawSendQueue.In <- frame.Bytes()
      infoBuf = new(bytes.Buffer)

      self.log("tick %d conns %d sessions %d", tick, len(self.conns), len(self.sessions))

      if self.maxConnNum > 0 && len(self.conns) == 0 { // client disconnected
        self.log("client disconnected")
        break LOOP
      }

    case conn := <-self.deadConnChan.Out:
      self.log("delete conn %d", conn.id)
      delete(self.conns, conn.id)

    case session := <-self.sessionStopNotify.Out:
      self.log("delete session %d", session.id)
      delete(self.sessions, session.id)
      self.stoppedSessions[session.id] = true

    case req := <-self.newSessionRequest:
      self.handleNewSessionRequest(req)

    case <-self.stop:
      break LOOP
    }
    tick++
  }

  // finalizer
  self.log("stop")
  self.closed = true
  // stop conns
  for _, conn := range self.conns {
    conn.Stop()
    delete(self.conns, conn.id)
  }
  // stop sessions
  for serial, session := range self.sessions {
    session.Stop()
    delete(self.sessions, serial)
    self.stoppedSessions[serial] = true
  }
  // notify 
  if self.stopNotify != nil {
    self.stopNotify <- self
  }
  // stop chans
  self.rawSendQueue.Stop()
  self.newConnChan.Stop()
  self.sessionStopNotify.Stop()
  self.deadConnChan.Stop()
  self.sendQueue.Stop()
  self.infoChan.Stop()
}

func (self *ConnPool) newSession(sessionId uint64) *Session {
  ret := make(chan *Session)
  self.newSessionRequest <- newSessionReq{ret, sessionId}
  return <-ret
}

func (self *ConnPool) handleNewSessionRequest(req newSessionReq) {
  sessionId, ret := req.sessionId, req.ret
  if self.stoppedSessions[sessionId] { // session stopped, do not create a new one
    ret <- nil
    return
  }
  session := newSession(sessionId, self)
  self.sessions[sessionId] = session
  session.stopNotify = self.sessionStopNotify
  session.infoChan = self.infoChan
  if self.newSessionChan != nil {
    *(self.newSessionChan) <- session
  }
  ret <- session
}

func (self *ConnPool) Stop() {
  close(self.stop)
}

func (self *ConnPool) log(f string, vars ...interface{}) {
  if DEBUG {
    colorp("34", "CONNPOOL " + f, vars...)
  }
}
