package main

// #cgo LDFLAGS: -lenet
// #include <enet/enet.h>
import "C"
import (
  "time"
  "net"
)

var startTime time.Time
var channelConn []net.Conn
var channelChan []chan []byte
var host *C.ENetHost

func init() {
  channelConn = make([]net.Conn, CHANNELS)
  channelChan = make([]chan []byte, CHANNELS)
  for i := 0; i < CHANNELS; i++ {
    channelChan[i] = make(chan []byte)
  }

  startTime = time.Now()
  C.enet_initialize()

  var address C.ENetAddress
  address.host = C.ENET_HOST_ANY
  address.port = HOST_PORT
  host = C.enet_host_create(&address, 2048, CHANNELS, 0, 0)
}
