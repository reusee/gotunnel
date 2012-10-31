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
var channelChan []chan *C.ENetEvent
var channelConn []net.Conn
var peer *C.ENetPeer

func init() {
  startTime = time.Now()

  idPool = make(chan C.enet_uint8, CHANNELS)
  channelChan = make([]chan *C.ENetEvent, CHANNELS)
  for i := 0; i < CHANNELS; i++ {
    idPool <- C.enet_uint8(i)
    channelChan[i] = make(chan *C.ENetEvent, 1024)
    go watchChan(channelChan[i])
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
      case C.ENET_EVENT_TYPE_RECEIVE:
        channelChan[event.channelID] <- &event

      case C.ENET_EVENT_TYPE_CONNECT:
      case C.ENET_EVENT_TYPE_DISCONNECT:
      case C.ENET_EVENT_TYPE_NONE:
      }
    }
  }()
}

func watchChan(c chan *C.ENetEvent) {
  for {
    event := <-c
    data := C.GoBytes(unsafe.Pointer(event.packet.data), C.int(event.packet.dataLength))
    C.enet_packet_destroy(event.packet)
    switch data[0] {
      case byte(1): // data packet
      dataLen := len(data) - 1
      decrypted := make([]byte, dataLen)
      xorSlice(data[1:], decrypted, dataLen, dataLen % 8)
      msg("[%d] server > %d\n", event.channelID, dataLen)
      channelConn[event.channelID].Write(decrypted)

      case byte(2): // end conn packet
      channelConn[event.channelID].Close()
    }
  }
}
