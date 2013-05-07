package gnet

import (
  "testing"
  "container/heap"
  "math/rand"
)

func TestPriorityQueue(t *testing.T) {
  queue := newPacketQueue()
  n := 1024
  perm := rand.Perm(n)
  for _, n := range perm {
    heap.Push(queue, &Packet{serial: uint32(n), data: []byte{}})
  }
  for i := 0; i < n; i++ {
    packet := heap.Pop(queue).(*Packet)
    if packet.serial != uint32(i) {
      t.Fatal("priority queue error")
    }
  }
}
