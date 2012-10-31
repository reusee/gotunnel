package main

// #cgo LDFLAGS: -lenet
// #include <enet/enet.h>
import "C"
import (
  "net"
  "log"
  "fmt"
  "encoding/binary"
  "io"
  "strconv"
  "hash/fnv"
  "bytes"
  "unsafe"
  "time"
)

var secret uint64
var uint64Keys []uint64
var byteKeys []byte

func init() {
  hasher := fnv.New64()
  hasher.Write([]byte(KEY))
  secret = hasher.Sum64()
  msg("secret %d\n", secret)

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
  msg("listening on %s\n", PORT)
  for {
    conn, err := ln.Accept()
    if err != nil {
      msg("accept error %v\n", err)
      continue
    }
    go handleConnection(conn)
  }
}

func handleConnection(conn net.Conn) {
  //defer conn.Close()
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
    msg("hostPort %s\n", hostPort)

    channelId := <-idPool
    channelConn[channelId] = conn
    defer func() {
      idPool <- channelId
      msg("return channel %d\n", channelId)
    }()
    msg("channel id %d\n", channelId)

    writeAck(conn, REP_SUCCEED)

    buf := new(bytes.Buffer)
    write(buf, byte(0))
    l := len(hostPort)
    encrypted := make([]byte, l)
    xorSlice([]byte(hostPort), encrypted, l, l % 8)
    write(buf, encrypted)
    packet := C.enet_packet_create(unsafe.Pointer(C.CString(string(buf.Bytes()))), C.size_t(buf.Len()), C.ENET_PACKET_FLAG_RELIABLE)
    C.enet_peer_send(peer, channelId, packet)

    buffer := make([]byte, 65535)
    for {
      n, err := conn.Read(buffer)
      if err != nil {
        conn.Close()
        break
      }
      msg("data len %d\n", n)
      l = n + 1
      data := make([]byte, l)
      data[0] = byte(1)
      xorSlice(buffer, data[1:], n, n % 8)
      packet := C.enet_packet_create(unsafe.Pointer(C.CString(string(data))), C.size_t(l), C.ENET_PACKET_FLAG_RELIABLE)
      C.enet_peer_send(peer, channelId, packet)
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

func msg(s string, vars ...interface{}) {
  fmt.Printf("%v " + s, append([]interface{}{time.Now().Sub(startTime)}, vars...)...)
}
