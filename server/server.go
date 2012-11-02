package main

import (
  "net"
  "log"
  "fmt"
  "encoding/binary"
  "io"
  "hash/fnv"
  "bytes"
  "sync/atomic"
  "time"
)

var secret uint64
var uint64Keys []uint64
var byteKeys []byte

var connCount int32

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

  go func() {
    ticker := time.NewTicker(time.Second * 1)
    for {
      <-ticker.C
      fmt.Printf("connections %d\n", connCount)
    }
  }()
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
  atomic.AddInt32(&connCount, int32(1))
  defer conn.Close()
  var hostPortLen uint8
  read(conn, &hostPortLen)
  encryptedHostPort := make([]byte, hostPortLen)
  read(conn, encryptedHostPort)
  hostPort := make([]byte, hostPortLen)
  xorSlice(encryptedHostPort, hostPort, int(hostPortLen), int(hostPortLen % 8))
  fmt.Printf("hostPort %s\n", hostPort)

  targetConn, err := net.Dial("tcp", string(hostPort))
  if err != nil {
    fmt.Printf("fail to connect %s\n", hostPort)
    return
  }
  defer targetConn.Close()

  go io.Copy(NewXorWriter(conn), targetConn)
  io.Copy(targetConn, NewXorReader(conn))
  atomic.AddInt32(&connCount, int32(-1))
}

func read(reader io.Reader, v interface{}) {
  binary.Read(reader, binary.BigEndian, v)
}

func write(writer io.Writer, v interface{}) {
  binary.Write(writer, binary.BigEndian, v)
}
