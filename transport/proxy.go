package gnet

import (
  "net"
  "sync"
  "time"
)

func (self *Session) ProxyTCP(conn *net.TCPConn, bufferSize int) {
  writeClosed := make(chan struct{})
  readClosed := make(chan struct{})
  go func() {
    <-writeClosed
    <-readClosed
    conn.Close()
    self.log("proxy closed")
  }()
  // to conn
  go func() {
    var once sync.Once
    for {
      select {
      case msg := <-self.Message:
        switch msg.Tag {
        case DATA:
          _, err := conn.Write(msg.Data)
          if err != nil {
            self.log("proxy write error %v", err)
            go once.Do(func() {
              <-time.After(time.Second * 5)
              conn.CloseWrite()
              close(writeClosed)
            })
            self.AbortRead()
          }
        case STATE:
          switch msg.State {
          case STATE_FINISH_SEND, STATE_ABORT_SEND:
            go once.Do(func() {
              <-time.After(time.Second * 5)
              conn.CloseWrite()
              close(writeClosed)
            })
          case STATE_ABORT_READ:
          case STATE_FINISH_READ:
          }
        }
      case <-self.Stopped:
        return
      }
    }
  }()
  // from conn
  for {
    buf := make([]byte, bufferSize)
    n, err := conn.Read(buf)
    if err != nil {
      conn.CloseRead()
      close(readClosed)
      self.FinishSend()
      return
    }
    self.Send(buf[:n])
  }
}
