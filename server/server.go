package main

import (
  "net"
  "log"
  "fmt"
  "encoding/binary"
  "io"
  "hash/fnv"
  "bytes"
)

var secret uint64
var uint64Keys []uint64
var byteKeys []byte

func init() {
  hasher := fnv.New64()
  hasher.Write([]byte(KEY))
  secret = hasher.Sum64()
  fmt.Printf("secret %d\n", secret)

  buf := new(bytes.Buffer)
  binary.Write(buf, binary.LittleEndian, secret)
  byteKeys = buf.Bytes()

  keys := byteKeys[:]
  for i := 0; i < 8; i++ {
    var key uint64
    binary.Read(buf, binary.LittleEndian, &key)
    uint64Keys = append(uint64Keys, key)
    keys = append(keys[1:], keys[0])
    buf = bytes.NewBuffer(keys)
  }
}

func main() {
  ln, err := net.Listen("tcp", PORT)
  if err != nil {
    log.Fatal("listen error on port %s\n", PORT)
  }
  fmt.Printf("listening on %s\n", PORT)
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

  go io.Copy(NewXorWriter(conn, secret), targetConn)
  io.Copy(targetConn, NewXorReader(conn, secret))
}

func read(reader io.Reader, v interface{}) {
  binary.Read(reader, binary.BigEndian, v)
}

func write(writer io.Writer, v interface{}) {
  binary.Write(writer, binary.BigEndian, v)
}
