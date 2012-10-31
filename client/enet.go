package main

// #cgo LDFLAGS: -lenet
// #include <enet/enet.h>
import "C"
import (
  "unsafe"
  "time"
  "log"
  "net"
)

var startTime time.Time

var idPool chan C.enet_uint8
var channelConn []net.Conn
var peer *C.ENetPeer

func init() {
  startTime = time.Now()

  idPool = make(chan C.enet_uint8, CHANNELS)
  for i := 0; i < CHANNELS; i++ {
    idPool <- C.enet_uint8(i)
  }
  channelConn = make([]net.Conn, CHANNELS)

  client := C.enet_host_create(nil, 2048, CHANNELS, 0, 0)
  var address C.ENetAddress
  C.enet_address_set_host(&address, C.CString(SERVER_HOST))
  address.port = SERVER_PORT

  var event C.ENetEvent
  peer = C.enet_host_connect(client, &address, CHANNELS, 0)
  if C.enet_host_service(client, &event, 3000) > 0 && event._type == C.ENET_EVENT_TYPE_CONNECT {
    msg("host connected\n")
  } else {
    C.enet_peer_reset(peer)
    log.Fatal("host connect fail")
  }

  go func() {
    for {
      var event C.ENetEvent
      C.enet_host_service(client, &event, 5)
      switch event._type {
      case C.ENET_EVENT_TYPE_CONNECT:
        msg("connected from %v %v\n", event.peer.address.host, event.peer.address.port)

      case C.ENET_EVENT_TYPE_RECEIVE:
        data := C.GoBytes(unsafe.Pointer(event.packet.data), C.int(event.packet.dataLength))
        C.enet_packet_destroy(event.packet)
        switch data[0] {
          case byte(1): // data packet
          dataLen := len(data) - 1
          decrypted := make([]byte, dataLen)
          xorSlice(data[1:], decrypted, dataLen, dataLen % 8)
          msg("receive data len %d channel %d\n", dataLen, event.channelID)
          channelConn[event.channelID].Write(decrypted)

          case byte(2): // end conn packet
          channelConn[event.channelID].Close()
        }

      case C.ENET_EVENT_TYPE_DISCONNECT:
      case C.ENET_EVENT_TYPE_NONE:
      }
    }
  }()
}
