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

  fromConn := make(chan []byte)
  go func() {
    for {
      buf := make([]byte, 4096)
      n, err := conn.Read(buf)
      fmt.Printf("conn read %d\n", n)
      if err != nil {
        fmt.Printf("err %v\n", err)
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
        fmt.Printf("receive %d\n", len(msg.Data))
        conn.Write(msg.Data)
      } else if msg.Tag == gnet.STATE && msg.State == gnet.STATE_STOP {
        fmt.Printf("stop\n")
        return
      }
    case data := <-fromConn:
      if data == nil {
        fmt.Printf("finish send\n")
        session.FinishSend()
      } else {
        fmt.Printf("send %d\n", len(data))
        session.Send(data)
      }
    }
  }

}
