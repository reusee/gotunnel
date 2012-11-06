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
  sessionCounter int64
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
      fmt.Printf("gotunnel: %d connections %d active sessions\n", connectionCounter, sessionCounter)
    }
  }()

  for {
    session := <-server.New
    atomic.AddInt64(&sessionCounter, int64(1))
    go handleSession(session)
  }
}

func handleSession(session *gnet.Session) {
  hostPort := string(<-session.Data)
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
  go func() {
    for {
      select {
      case data := <-session.Data:
        conn.Write(data)
      case state := <-session.State:
        if state == gnet.STATE_FINISH_SEND {
          atomic.AddInt64(&sessionCounter, int64(-1))
          return
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
    session.Send(buf[:n])
  }
}
