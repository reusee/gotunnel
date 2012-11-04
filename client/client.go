package main

import (
  "bytes"
  "encoding/binary"
  "flag"
  "fmt"
  "hash/fnv"
  "io"
  "log"
  "net"
  "os"
  "strconv"
  "sync/atomic"
)

var secret uint64
var uint64Keys []uint64
var byteKeys []byte
var connCount int32

var gKey, gPort, gServer string

func usage() {
  flag.PrintDefaults()
  os.Exit(2)
}

func parseArgs() {
  flag.StringVar(&gKey, "key", "", "key for encryption")
  flag.StringVar(&gPort, "port", "127.0.0.1:8809", "listen on addr:port")
  flag.StringVar(&gServer, "server", "", "connect to server addr:port")
  flag.Parse()

  if gKey == "" || gPort == "" || gServer == "" { usage() }
}

func init() {
  parseArgs()

  hasher := fnv.New64()
  hasher.Write([]byte(gKey))
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
  ln, err := net.Listen("tcp", gPort)
  if err != nil {
    log.Fatal("listen error on port %s\n", gPort)
  }
  fmt.Printf("listening on %s\n", gPort)
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
  var ver, nMethods byte

  read(conn, &ver)
  read(conn, &nMethods)
  methods := make([]byte, nMethods)
  read(conn, methods)
  write(conn, VERSION)
  if ver != VERSION || nMethods != byte(1) || methods[0] != METHOD_NOT_REQUIRED {
    write(conn, METHOD_NO_ACCEPTABLE)
  } else {
    write(conn, METHOD_NOT_REQUIRED)
  }

  var cmd, reserved, addrType byte
  read(conn, &ver)
  read(conn, &cmd)
  read(conn, &reserved)
  read(conn, &addrType)
  if ver != VERSION {
    return
  }
  if reserved != RESERVED {
    return
  }
  if addrType != ADDR_TYPE_IP && addrType != ADDR_TYPE_DOMAIN {
    writeAck(conn, REP_ADDRESS_TYPE_NOT_SUPPORTED)
    return
  }

  var address []byte
  if addrType == ADDR_TYPE_IP {
    address = make([]byte, 4)
    read(conn, address)
  } else if addrType == ADDR_TYPE_DOMAIN {
    var domainLength byte
    read(conn, &domainLength)
    address = make([]byte, domainLength)
    read(conn, address)
  }
  var port uint16
  read(conn, &port)

  switch cmd {
  case CMD_CONNECT:
    var hostPort string
    if addrType == ADDR_TYPE_IP {
      ip := net.IPv4(address[0], address[1], address[2], address[3])
      hostPort = net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
    } else if addrType == ADDR_TYPE_DOMAIN {
      hostPort = net.JoinHostPort(string(address), strconv.Itoa(int(port)))
    }
    fmt.Printf("hostPort %s\n", hostPort)

    serverConn, err := net.Dial("tcp", gServer)
    if err != nil {
      fmt.Printf("server connect fail %v\n", err)
      writeAck(conn, REP_SERVER_FAILURE)
      return
    }
    defer serverConn.Close()

    writeAck(conn, REP_SUCCEED)

    hostPortLen := len(hostPort)
    write(serverConn, byte(hostPortLen))
    encryptedHostPort := make([]byte, hostPortLen)
    xorSlice([]byte(hostPort), encryptedHostPort, hostPortLen, hostPortLen % 8)
    write(serverConn, encryptedHostPort)

    go io.Copy(conn, NewXorReader(serverConn))
    io.Copy(NewXorWriter(serverConn), conn)

  case CMD_BIND:
  case CMD_UDP_ASSOCIATE:
  default:
    writeAck(conn, REP_COMMAND_NOT_SUPPORTED)
    return
  }

  atomic.AddInt32(&connCount, int32(-1))
}

func writeAck(conn net.Conn, reply byte) {
  write(conn, VERSION)
  write(conn, reply)
  write(conn, RESERVED)
  write(conn, ADDR_TYPE_IP)
  write(conn, [4]byte{0, 0, 0, 0})
  write(conn, uint16(0))
}

func read(reader io.Reader, v interface{}) {
  binary.Read(reader, binary.BigEndian, v)
}

func write(writer io.Writer, v interface{}) {
  binary.Write(writer, binary.BigEndian, v)
}
