package gnet

import (
  "container/heap"
  "container/list"
)

type PacketQueue struct {
  *list.List
}

func newPacketQueue() *PacketQueue {
  packetQueue := PacketQueue{list.New()}
  heap.Init(&packetQueue)
  return &packetQueue
}

func at(l PacketQueue, i int) *list.Element {
  e := l.Front()
  n := 0
  for n < i {
    e = e.Next()
    n++
  }
  return e
}

func (self PacketQueue) Less(i, j int) bool {
  return at(self, i).Value.(*Packet).serial < at(self, j).Value.(*Packet).serial
}

func (self PacketQueue) Swap(i, j int) {
  ei := at(self, i)
  ej := at(self, j)
  ei.Value, ej.Value = ej.Value, ei.Value
  ei.Value.(*Packet).index = i
  ej.Value.(*Packet).index = j
}

func (self *PacketQueue) Push(x interface{}) {
  item := x.(*Packet)
  item.index = self.Len()
  self.PushBack(item)
}

func (self *PacketQueue) Pop() interface{} {
  elem := self.Back()
  item := elem.Value.(*Packet)
  item.index = -1
  self.Remove(elem)
  return item
}
