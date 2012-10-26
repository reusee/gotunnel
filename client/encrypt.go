package main

import (
  "bytes"
  "io"
  "encoding/binary"
)

func NewXorWriter(writer io.Writer, secret uint64) *Writer {
  self := &Writer{
    writer: writer,
    keyIndex: 0,
  }
  buf := new(bytes.Buffer)
  binary.Write(buf, binary.LittleEndian, secret)
  self.keys = buf.Bytes()
  return self
}

type Writer struct {
  writer io.Writer
  keyIndex int
  keys []byte
}

func (self *Writer) Write(p []byte) (n int, err error) {
  buf := make([]byte, len(p))
  for i, b := range p {
    buf[i] = b ^ self.keys[self.keyIndex]
    self.keyIndex++
    if self.keyIndex == 4 {
      self.keyIndex = 0
    }
  }
  return self.writer.Write(buf)
}

func NewXorReader(reader io.Reader, secret uint64) *Reader {
  self := &Reader{
    reader: reader,
    keyIndex: 0,
  }
  buf := new(bytes.Buffer)
  binary.Write(buf, binary.LittleEndian, secret)
  self.keys = buf.Bytes()
  return self
}

type Reader struct {
  reader io.Reader
  keyIndex int
  keys []byte
}

func (self *Reader) Read(p []byte) (n int, err error) {
  buf := make([]byte, len(p))
  n, err = self.reader.Read(buf)
  for i := 0; i < n; i++ {
    p[i] = buf[i] ^ self.keys[self.keyIndex]
    self.keyIndex++
    if self.keyIndex == 4 {
      self.keyIndex = 0
    }
  }
  return
}
