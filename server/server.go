package main

import (
  gnet "../gnet"
  "log"
  "fmt"
  "net"
  "sync/atomic"
  "time"
)

const LIMIT_FACTOR = 256

var (
  connectionCounter int64
)

func main() {
  server, err := gnet.NewServer(PORT, KEY)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("listening on %s\n", PORT)

  for {
    session := <-server.New
    go handleSession(session)
  }
}

func handleSession(session *gnet.Session) {
  hostPort := string((<-session.Message).Data)
  fmt.Printf("hostPort: %s\n", hostPort)
  conn, err := net.Dial("tcp", hostPort)
  atomic.AddInt64(&connectionCounter, int64(1))
  if err != nil {
    session.Send([]byte{0})
    return
  } else {
    session.Send([]byte{1})
  }
  defer func() {
    conn.Close()
    atomic.AddInt64(&connectionCounter, int64(-1))
  }()

  fromConn := make(chan []byte)
  go func() {
    for {
      buf := make([]byte, 4096)
      delta := session.BytesSent - session.RemoteBytesRead
      sleep := time.Duration(delta * 1000 / (LIMIT_FACTOR * 1024 * 1024)) * time.Millisecond
      time.Sleep(sleep)
      n, err := conn.Read(buf)
      if err != nil {
        fromConn <- nil
        return
      }
      fromConn <- buf[:n]
    }
  }()

  for {
    select {
    case msg := <-session.Message:
      if msg.Tag == gnet.DATA {
        if _, err := conn.Write(msg.Data); err != nil {
          session.FinishRead()
          return
        }
      } else if msg.Tag == gnet.STATE && msg.State == gnet.STATE_STOP {
        return
      }
    case data := <-fromConn:
      if data == nil {
        session.FinishSend()
      } else {
        session.Send(data)
      }
    }
  }

}
