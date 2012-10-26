package main

import (
  "io"
  "reflect"
  "unsafe"
  "fmt"
)

func NewXorWriter(writer io.Writer, secret uint64) *Writer {
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
  j := 0
  if l >= 8 {
    bufU64Slice := getUint64Slice(buf)
    pU64Slice := getUint64Slice(p)
    for i, e := range pU64Slice {
      bufU64Slice[i] = e ^ uint64Keys[self.keyIndex]
      j += 8
    }
  }
  if l % 8 > 0 {
    for i := 0; i < l % 8; i++ {
      buf[i + j] = p[i + j] ^ byteKeys[self.keyIndex]
      self.keyIndex++
      if self.keyIndex == 8 {
        self.keyIndex = 0
      }
    }
  }
  return self.writer.Write(buf)
}

func NewXorReader(reader io.Reader, secret uint64) *Reader {
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
  j := 0
  if n >= 8 {
    pU64Slice := getUint64Slice(p)
    bufU64Slice := getUint64Slice(buf)
    for i := 0; i < n / 8; i++ {
      pU64Slice[i] = bufU64Slice[i] ^ uint64Keys[self.keyIndex]
      j += 8
    }
  }
  if n % 8 > 0 {
    for i := 0; i < n % 8; i++ {
      p[i + j] = buf[i + j] ^ byteKeys[self.keyIndex]
      self.keyIndex++
      if self.keyIndex == 8 {
        self.keyIndex = 0
      }
    }
  }
  return
}

func getUint64Slice(s []byte) []uint64 {
  u64Slice := make([]uint64, 0, 0)
  header := (*reflect.SliceHeader)(unsafe.Pointer(&u64Slice))
  header.Data = (*reflect.SliceHeader)(unsafe.Pointer(&s)).Data
  header.Len = len(s) / 8
  header.Cap = len(s) / 8
  return u64Slice
}
