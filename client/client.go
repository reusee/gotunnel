package main

import (
  "net"
  "log"
  "fmt"
  "encoding/binary"
  "io"
  "strconv"
  "hash/fnv"
)

const BUF_SIZE = 20480

var (
  port = ":8808"
  serverHostPort = "127.0.0.1:38808"
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

  if cmd == CMD_CONNECT {
    var hostPort string
    if addrType == ADDR_TYPE_IP {
      ip := net.IPv4(address[0], address[1], address[2], address[3])
      hostPort = net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
    } else if addrType == ADDR_TYPE_DOMAIN {
      hostPort = net.JoinHostPort(string(address), strconv.Itoa(int(port)))
    }
    fmt.Printf("hostPort %s\n", hostPort)

    serverConn, err := net.Dial("tcp", serverHostPort)
    if err != nil {
      fmt.Printf("server connect fail %v\n", err)
      writeAck(conn, REP_SERVER_FAILURE)
      return
    }
    defer serverConn.Close()

    writeAck(conn, REP_SUCCEED)

    write(serverConn, secret)
    write(serverConn, byte(len(hostPort)))
    write(serverConn, []byte(hostPort))

    go forwardToServer(conn, serverConn)
    forwardToConn(serverConn, conn)

  } else if cmd == CMD_BIND {
  } else if cmd == CMD_UDP_ASSOCIATE {
  } else {
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

func forwardToServer(in net.Conn, out net.Conn) {
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

func forwardToConn(in net.Conn, out net.Conn) {
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
