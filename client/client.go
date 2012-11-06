package main

import (
  "net"
  "log"
  "fmt"
  "encoding/binary"
  "io"
  "strconv"
  gnet "../gnet"
  "sync/atomic"
  "time"
)

var (
  client *gnet.Client
  sessionCounter int64
  connectionCounter int64
)

func main() {
  var err error
  client, err = gnet.NewClient(SERVER, KEY, 128)
  if err != nil {
    log.Fatal(err)
  }

  ln, err := net.Listen("tcp", PORT)
  if err != nil {
    log.Fatal("listen error on port %s\n", PORT)
  }
  fmt.Printf("listening on %s\n", PORT)

  go func() {
    heartBeat := time.NewTicker(time.Second * 1)
    for {
      <-heartBeat.C
      fmt.Printf("gotunnel: %d connections %d active sessions\n", connectionCounter, sessionCounter)
    }
  }()

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
  atomic.AddInt64(&connectionCounter, int64(1))
  defer func() {
    conn.Close()
    atomic.AddInt64(&connectionCounter, int64(-1))
  }()

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

    writeAck(conn, REP_SUCCEED)

    session := client.NewSession()
    atomic.AddInt64(&sessionCounter, int64(1))
    session.Send([]byte(hostPort))
    ret := (<-session.Data)[0]
    if ret != byte(1) {
      return
    }
    fmt.Printf("hostPort %s %v\n", hostPort, ret)

    // send to client
    abort := make(chan bool)
    go func() {
      for {
        select {
        case data := <-session.Data:
          conn.Write(data)
        case <-abort:
          return
        }
      }
    }()

    // read from client and send to server
    buf := make([]byte, 4096)
    for {
      n, err := conn.Read(buf)
      if err != nil {
        session.Abort()
        atomic.AddInt64(&sessionCounter, int64(-1))
        abort <- true
        break
      }
      session.Send(buf[:n])
    }

  case CMD_BIND:
  case CMD_UDP_ASSOCIATE:
  default:
    writeAck(conn, REP_COMMAND_NOT_SUPPORTED)
    return
  }
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

func msg(f string, vars ...interface{}) {
}
