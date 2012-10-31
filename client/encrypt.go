package main

import (
  "reflect"
  "unsafe"
)

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
