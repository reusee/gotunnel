package gnet

import (
  "encoding/binary"
  "io"
  "net"
  "bytes"
)

func readFrame(conn *net.TCPConn) ([]byte, error) {
  var length uint32
  err := binary.Read(conn, binary.BigEndian, &length)
  if err != nil {
    return nil, err
  }
  payload := make([]byte, length)
  _, err = io.ReadFull(conn, payload)
  if err != nil {
    return nil, err
  }
  return payload, nil
}

func writeFrame(conn *net.TCPConn, payload []byte) error {
  err := binary.Write(conn, binary.BigEndian, uint32(len(payload)))
  if err != nil {
    return err
  }
  _, err = conn.Write(payload)
  if err != nil {
    return err
  }
  return nil
}

func parseSessionPacket(packet []byte, byteKeys []byte, uint64Keys []uint64) ([]byte, uint64) {
  var sessionId uint64
  binary.Read(bytes.NewReader(packet[:8]), binary.BigEndian, &sessionId)
  dataLen := len(packet) - 8
  data := packet[8:]
  xorSlice(data, dataLen, byteKeys, uint64Keys)
  return data, sessionId
}

func assembleSessionPacket(sessionId uint64, data []byte, byteKeys []byte, uint64Keys []uint64) []byte {
  buf := new(bytes.Buffer)
  buf.Write([]byte{PACKET_TYPE_SESSION}) // packet type
  binary.Write(buf, binary.BigEndian, sessionId) // session id
  dataLen := len(data)
  xorSlice(data, dataLen, byteKeys, uint64Keys)
  buf.Write(data) // data
  return buf.Bytes()
}
