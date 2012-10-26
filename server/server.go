package main

import (
  "net"
  "log"
  "fmt"
  "encoding/binary"
  "io"
  "hash/fnv"
)

const BUF_SIZE = 20480

var (
  key = "abcdar"
  secret uint64
)

func init() {
  hasher := fnv.New64()
  hasher.Write([]byte(key))
  secret = hasher.Sum64()
  fmt.Printf("secret %d\n", secret)
}

func main() {
  port := ":38808"
  ln, err := net.Listen("tcp", port)
  if err != nil {
    log.Fatal("listen error on port %s\n", port)
  }
  fmt.Printf("listening on %s\n", port)
  for {
    conn, err := ln.Accept()
    if err != nil {
      fmt.Printf("accept error %v\n", err)
      continue
    }
    go handleConnection(conn)
  }
}

func handleConnection(conn net.Conn) {
  defer conn.Close()
  var connSecret uint64
  read(conn, &connSecret)
  if connSecret != secret {
    fmt.Printf("secret not match\n")
    return
  }
  var hostPortLen uint8
  read(conn, &hostPortLen)
  hostPort := make([]byte, hostPortLen)
  read(conn, hostPort)
  fmt.Printf("hostPort %s\n", hostPort)

  targetConn, err := net.Dial("tcp", string(hostPort))
  if err != nil {
    fmt.Printf("fail to connect %s\n", hostPort)
    return
  }
  defer targetConn.Close()

  go forwardToHost(conn, targetConn)
  forwardToClient(targetConn, conn)
}

func forwardToHost(in net.Conn, out net.Conn) {
  buf := make([]byte, BUF_SIZE)
  for {
    n, err := in.Read(buf)
    if err == io.EOF {
      break
    } else if err != nil {
      break
    }
    fmt.Printf("%d\n", n)
    out.Write(buf[:n])
  }
}

func forwardToClient(in net.Conn, out net.Conn) {
  buf := make([]byte, BUF_SIZE)
  for {
    n, err := in.Read(buf)
    if err == io.EOF {
      break
    } else if err != nil {
      break
    }
    fmt.Printf("%d\n", n)
    out.Write(buf[:n])
  }
}

func read(reader io.Reader, v interface{}) {
  binary.Read(reader, binary.BigEndian, v)
}

func write(writer io.Writer, v interface{}) {
  binary.Write(writer, binary.BigEndian, v)
}
