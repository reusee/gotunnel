package main

import (
  tp "../transport"
  "log"
  "fmt"
  "net"
  "runtime"
  "time"
)

const LIMIT_FACTOR = 256

func main() {
  runtime.GOMAXPROCS(3)

  server, err := tp.NewServer(PORT, KEY)
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

func handleSession(session *tp.Session) {
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

  // start forward
  session.ProxyTCP(conn, 4096)
}
