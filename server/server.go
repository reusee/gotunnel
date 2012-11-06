package main

import (
  gnet "../gnet"
  "log"
  "fmt"
  "net"
)

func main() {
  server, err := gnet.NewServer(PORT, KEY)
  if err != nil {
    log.Fatal(err)
  }
  for {
    session := <-server.New
    go handleSession(session)
  }
}

func handleSession(session *gnet.Session) {
  hostPort := string(<-session.Data)
  fmt.Printf("hostPort: %s\n", hostPort)
  conn, err := net.Dial("tcp", hostPort)
  if err != nil {
    session.Send([]byte{0})
    return
  } else {
    session.Send([]byte{1})
  }
  defer conn.Close()

  end := make(chan bool)
  go func() {
    buf := make([]byte, 4096)
    for {
      n, err := conn.Read(buf)
      if err != nil {
        end <- true
        break
      }
      session.Send(buf[:n])
    }
  }()

  for {
    select {
    case data := <-session.Data:
      conn.Write(data)
    case <-end:
      return
    }
  }
}
