package main

import (
  "net"
  "log"
  "fmt"
  "encoding/binary"
  "io"
  "strconv"
  gnet "../gnet"
  "time"
  "runtime"
  "math/rand"
)

var (
  client *gnet.Client
)

func main() {
  var err error
  client, err = gnet.NewClient(SERVER, KEY, 32)
  if err != nil {
    log.Fatal(err)
  }

  addr, err := net.ResolveTCPAddr("tcp", PORT)
  if err != nil {
    log.Fatal(err)
  }
  ln, err := net.ListenTCP("tcp", addr)
  if err != nil {
    log.Fatal("listen error on port %s\n", PORT)
  }
  fmt.Printf("listening on %s\n", PORT)

  go func() {
    heartBeat := time.NewTicker(time.Second * 5)
    for {
      <-heartBeat.C

      if client.Closed {
        ln.Close()
        return
      }

      fmt.Printf("gotunnel client: sent %d bytes, read %d bytes, %d goroutines\n",
        client.BytesSent,
        client.BytesRead,
        runtime.NumGoroutine())

    }
  }()

  for {
    conn, err := ln.AcceptTCP()
    if err != nil {
      if client.Closed {
        return
      }
      fmt.Printf("accept error %v\n", err)
      continue
    }
    go handleConnection(conn)
  }
}

func handleConnection(conn *net.TCPConn) {
  var ver, nMethods byte

  // handshake
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

  // request
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

  var hostPort string
  if addrType == ADDR_TYPE_IP {
    ip := net.IPv4(address[0], address[1], address[2], address[3])
    hostPort = net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
  } else if addrType == ADDR_TYPE_DOMAIN {
    hostPort = net.JoinHostPort(string(address), strconv.Itoa(int(port)))
  }

  switch cmd {
  case CMD_CONNECT:
    handleConnect(hostPort, conn)
  case CMD_BIND:
    writeAck(conn, REP_COMMAND_NOT_SUPPORTED)
  case CMD_UDP_ASSOCIATE:
    writeAck(conn, REP_COMMAND_NOT_SUPPORTED)
  default:
    writeAck(conn, REP_COMMAND_NOT_SUPPORTED)
    return
  }
}

func handleConnect(hostPort string, conn *net.TCPConn) {
  writeAck(conn, REP_SUCCEED)

  uid := rand.Int63()
  info := func(f string, vars ...interface{}) {
    if gnet.DEBUG {
      fmt.Printf(fmt.Sprintf("%d %s\n", uid, f), vars...)
    }
  }

  info("hostPort %s", hostPort)

  session := client.NewSession()
  session.Send([]byte(hostPort))
  select {
  case msg := <-session.Message:
    if msg.Tag != gnet.DATA {
      info("get non-data msg")
      return
    }
    retCode := msg.Data[0]
    if retCode != byte(1) {
      info("remote dial failed")
      return
    }
  case <-session.Stopped:
    info("session stopped")
    return
  }

  // start forward
  session.ProxyTCP(conn, 4096)
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
