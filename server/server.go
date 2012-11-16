package main

import (
  gnet "../gnet"
  "log"
  "fmt"
  "net"
  "runtime"
  "time"
)

const LIMIT_FACTOR = 256

func main() {
  server, err := gnet.NewServer(PORT, KEY)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("listening on %s\n", PORT)

  go func() { // heartBeat
    heartBeat := time.NewTicker(time.Second * 2)
    for _ = range heartBeat.C {
      fmt.Printf("sent %d bytes, read %d bytes, %d goroutines\n",
        server.BytesSent,
        server.BytesRead,
        runtime.NumGoroutine())
    }
  }()

  for {
    session := <-server.New
    go handleSession(session)
  }
}

func handleSession(session *gnet.Session) {
  var hostPort string

  select {
  case msg := <-session.Message:
    hostPort = string(msg.Data)
  case <-session.Stopped:
    return
  }

  fmt.Printf("hostPort: %s\n", hostPort)
  addr, err := net.ResolveTCPAddr("tcp", hostPort)
  if err != nil {
    return
  }
  conn, err := net.DialTCP("tcp", nil, addr)
  if err != nil {
    session.Send([]byte{0})
    return
  } else {
    session.Send([]byte{1})
  }

  go func() {
    buf := make([]byte, 4096)
    for {
      n, err := conn.Read(buf)
      if err != nil { // target close send
        conn.CloseRead()
        session.FinishSend()
        return
      }
      session.Send(buf[:n])
    }
  }()

  LOOP:
  for {
    select {
    case msg := <-session.Message:
      switch msg.Tag {
      case gnet.DATA:
        conn.Write(msg.Data)
      case gnet.STATE:
        if msg.State == gnet.STATE_FINISH_SEND {
          conn.CloseWrite()
          session.FinishRead()
        }
      }
    case <-session.Stopped:
      break LOOP
    }
  }

}
