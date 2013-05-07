package gnet

import (
  "testing"
  "fmt"
  "bytes"
)

func TestNew(t *testing.T) {
  p("Start TestNew\n")
  server, err := NewServer(":8888", "abcd")
  if err != nil {
    t.Fatal(err)
  }
  defer server.Stop()

  client, err := NewClient("localhost:8888", "abcd", 1)
  if err != nil {
    t.Fatal(err)
  }
  defer client.Stop()

  n := 2000

  go func() {
    session := client.NewSession()
    for i := 0; i < n; i++ {
      session.C.In <- []byte(fmt.Sprintf("%d", i))
      <-session.Message
    }
  }()

  session := <-server.New
  for i := 0; i < n; i++ {
    data := (<-session.Message).Data
    p("%s\n", data)
    expected := []byte(fmt.Sprintf("%d", i))
    if bytes.Compare(data, expected) != 0 {
      t.Fatal("wrong seq")
    }
    session.Send(data)
  }
}

func TestSessionAbort(t *testing.T) {
  p("Start TestSessionAbort\n")
  server, err := NewServer(":8700", "abc")
  if err != nil {
    t.Fatal(err)
  }
  defer server.Stop()

  client, err := NewClient("localhost:8700", "abc", 1)
  if err != nil {
    t.Fatal(err)
  }
  defer client.Stop()

  clientSession := client.NewSession()

  go func() {
    session := <-server.New
    c := 0
    for {
      msg := <-session.Message
      switch msg.Tag {
      case DATA:
        data := msg.Data
        session.Send(data)
        c++
        if c > 100 {
          clientSession.AbortSend()
        }
      case STATE:
        if msg.State == STATE_ABORT_SEND {
          fmt.Printf("\nclient abort send. quit\n")
          session.AbortRead()
          return
        }
      }
    }
  }()

  for i := 0; i < 200; i++ {
    clientSession.Send([]byte(fmt.Sprintf("%d", i)))
  }
  var data []byte
  for {
    msg := <-clientSession.Message
    switch msg.Tag {
    case DATA:
      data = msg.Data
      p("%s ", data)
    case STATE:
      if msg.State == STATE_ABORT_READ {
        fmt.Printf("\nserver abort read. quit. last echo %s\n", data)
        return
      }
    }
  }
}

func TestSessionFinish(t *testing.T) {
  p("Start TestSessionFinish\n")
  server, err := NewServer(":8710", "abc")
  if err != nil {
    t.Fatal(err)
  }
  defer func() {
    p("server sent %d bytes, read %d bytes\n", server.BytesSent, server.BytesRead)
    server.Stop()
  }()

  client, err := NewClient("localhost:8710", "abc", 1)
  if err != nil {
    t.Fatal(err)
  }
  defer func() {
    p("client sent %d bytes, read %d bytes\n", client.BytesSent, client.BytesRead)
    client.Stop()
  }()

  clientSession := client.NewSession()

  go func() {
    session := <-server.New
    for {
      msg := <-session.Message
      switch msg.Tag {
      case DATA:
        data := msg.Data
        session.Send(data)
      case STATE:
        if msg.State == STATE_FINISH_SEND {
          fmt.Printf("\nclient finish send. quit\n")
          session.FinishSend()
          return
        }
      }
    }
  }()

  n := 2000
  for i := 0; i < n; i++ {
    clientSession.Send([]byte(fmt.Sprintf("%d", i)))
  }
  var data []byte
  clientSession.FinishSend()
  for {
    msg := <-clientSession.Message
    switch msg.Tag {
    case DATA:
      data = msg.Data
      p("%s ", data)
    case STATE:
      if msg.State == STATE_FINISH_SEND {
        fmt.Printf("\nserver finish send. quit. last echo %s\n", data)
        return
      }
    }
  }
  if !bytes.Equal(data, []byte(fmt.Sprintf("%d", n - 1))) {
    t.Fail()
  }
}
