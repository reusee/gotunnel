package main

import (
  "io"
  "reflect"
  "unsafe"
)

func NewXorWriter(writer io.Writer) *Writer {
  self := &Writer{
    writer: writer,
    keyIndex: 0,
  }
  return self
}

type Writer struct {
  writer io.Writer
  keyIndex int
}

func (self *Writer) Write(p []byte) (n int, err error) {
  l := len(p)
  buf := make([]byte, l)
  self.keyIndex = xorSlice(p, buf, l, self.keyIndex)
  return self.writer.Write(buf)
}

func NewXorReader(reader io.Reader) *Reader {
  self := &Reader{
    reader: reader,
    keyIndex: 0,
  }
  return self
}

type Reader struct {
  reader io.Reader
  keyIndex int
}

func (self *Reader) Read(p []byte) (n int, err error) {
  l := len(p)
  buf := make([]byte, l)
  n, err = self.reader.Read(buf)
  self.keyIndex = xorSlice(buf, p, n, self.keyIndex)
  return
}

func xorSlice(from []byte, to []byte, n int, keyIndex int) int {
  j := 0
  if n >= 8 {
    toU64Slice := getUint64Slice(to)
    fromU64Slice := getUint64Slice(from)
    for i := 0; i < n / 8; i++ {
      toU64Slice[i] = fromU64Slice[i] ^ uint64Keys[keyIndex]
      j += 8
    }
  }
  for j < n {
    to[j] = from[j] ^ byteKeys[keyIndex]
    keyIndex++
    if keyIndex == 8 {
      keyIndex = 0
    }
    j++
  }
  return keyIndex
}

func getUint64Slice(s []byte) []uint64 {
  u64Slice := make([]uint64, 0, 0)
  header := (*reflect.SliceHeader)(unsafe.Pointer(&u64Slice))
  header.Data = (*reflect.SliceHeader)(unsafe.Pointer(&s)).Data
  header.Len = len(s) / 8
  header.Cap = len(s) / 8
  return u64Slice
}
