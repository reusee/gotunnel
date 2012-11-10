package main

import (
  gnet "../gnet"
  "log"
  "fmt"
  "net"
  "time"
  "sync/atomic"
)

var (
  connectionCounter int64
)

func main() {
  server, err := gnet.NewServer(PORT, KEY)
  if err != nil {
    log.Fatal(err)
  }

  go func() {
    heartBeat := time.NewTicker(time.Second * 1)
    for {
      <-heartBeat.C
      fmt.Printf("gotunnel: %d connections\n", connectionCounter)
    }
  }()

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

  // read from client and send to target
  clientAbort := false
  go func() {
    for {
      msg := <-session.Message
      switch msg.Tag {
      case gnet.DATA:
        conn.Write(msg.Data)
      case gnet.STATE:
        if msg.State == gnet.STATE_FINISH_SEND { // client finish send
          return
        } else if msg.State == gnet.STATE_ABORT_READ || msg.State == gnet.STATE_ABORT_SEND {
          session.Abort()
          clientAbort = true
        }
      }
    }
  }()

  // read from target
  buf := make([]byte, 4096)
  for {
    n, err := conn.Read(buf)
    if err != nil {
      session.FinishSend()
      break
    }
    if clientAbort {
      break
    }
    session.Send(buf[:n])
  }
}
