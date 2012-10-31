package main

// #cgo LDFLAGS: -lenet
// #include <enet/enet.h>
import "C"
import (
  "net"
  "fmt"
  "encoding/binary"
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
  for {
    var event C.ENetEvent
    C.enet_host_service(host, &event, 5)
    switch event._type {
    case C.ENET_EVENT_TYPE_CONNECT:
      msg("peer connected %v %v\n", event.peer.address.host, event.peer.address.port)

    case C.ENET_EVENT_TYPE_RECEIVE:
      go handlePacket(&event)

    case C.ENET_EVENT_TYPE_DISCONNECT:
    case C.ENET_EVENT_TYPE_NONE:
    }
  }
}

func handlePacket(event *C.ENetEvent) {
  data := C.GoBytes(unsafe.Pointer(event.packet.data), C.int(event.packet.dataLength))
  C.enet_packet_destroy(event.packet)
  chanId := int(event.channelID)
  switch data[0] {
    case byte(0): // hostPort
      hostPortLen := len(data) - 1
      hostPort := make([]byte, hostPortLen)
      xorSlice(data[1:], hostPort, int(hostPortLen), int(hostPortLen % 8))
      msg("[%d] host %s\n", event.channelID, hostPort)

      conn, err := net.Dial("tcp", string(hostPort))
      if err != nil {
        msg("[%d] connect fail\n", event.channelID, hostPort)
        return
      }
      msg("[%d] connected %s\n", event.channelID, hostPort)
      channelConn[chanId] = conn

      exit := make(chan bool)
      go func() {
        for {
          select {
          case <-exit:
            break
          case data := <-channelChan[chanId]:
            conn.Write(data)
          }
        }
      }()
      buf := make([]byte, 65535)
      for {
        n, err := conn.Read(buf)
        if err != nil {
          break
        }
        msg("[%d] target > %d\n", event.channelID, n)
        l := n + 1
        data := make([]byte, l)
        data[0] = byte(1)
        xorSlice(buf, data[1:], n, n % 8)
        packet := C.enet_packet_create(unsafe.Pointer(C.CString(string(data))), C.size_t(l), C.ENET_PACKET_FLAG_RELIABLE)
        C.enet_peer_send(event.peer, event.channelID, packet)
      }
      defer func() {
        conn.Close()
        exit <- true
        data := []byte{2}
        packet := C.enet_packet_create(unsafe.Pointer(C.CString(string(data))), C.size_t(1), C.ENET_PACKET_FLAG_RELIABLE)
        C.enet_peer_send(event.peer, event.channelID, packet)
      }()

    case byte(1): // data packet
      dataLen := len(data) - 1
      msg("[%d] client > %d\n", event.channelID, dataLen)
      decrypted := make([]byte, dataLen)
      xorSlice(data[1:], decrypted, dataLen, dataLen % 8)
      channelChan[chanId] <- decrypted
  }
}

func msg(s string, vars ...interface{}) {
  fmt.Printf("%v " + s, append([]interface{}{time.Now().Sub(startTime)}, vars...)...)
}
