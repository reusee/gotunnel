package gnet

import (
  "reflect"
  "unsafe"
  "hash/fnv"
  "bytes"
  "encoding/binary"
)

func calculateKeys(keyStr string) (byteKeys []byte, uint64Keys []uint64) {
  hasher := fnv.New64()
  hasher.Write([]byte(keyStr))
  secret := hasher.Sum64()

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
  return
}

func xorSlice(s []byte, n int, byteKeys []byte, uint64Keys []uint64) {
  keyIndex := n % 8
  j := 0
  if n >= 8 {
    u64Slice := getUint64Slice(s)
    for i := 0; i < n / 8; i++ {
      u64Slice[i] = u64Slice[i] ^ uint64Keys[keyIndex]
      j += 8
    }
  }
  for j < n {
    s[j] = s[j] ^ byteKeys[keyIndex]
    keyIndex++
    if keyIndex == 8 {
      keyIndex = 0
    }
    j++
  }
}

func getUint64Slice(s []byte) []uint64 {
  u64Slice := make([]uint64, 0, 0)
  header := (*reflect.SliceHeader)(unsafe.Pointer(&u64Slice))
  header.Data = (*reflect.SliceHeader)(unsafe.Pointer(&s)).Data
  header.Len = len(s) / 8
  header.Cap = len(s) / 8
  return u64Slice
}
