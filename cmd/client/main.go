package main

import (
  "net"
  "log"
  "encoding/gob"

  "godrone"
)

func main() {
  log.Println("Starting the client.")
  conn, err := net.Dial("tcp", "192.168.1.1:666")
  if err != nil {
    panic(err)
  }

  encoder := gob.NewEncoder(conn)
  motors := &godrone.MotorPWM{[4]float64{0.1,0.1,0.1,0.1}}
  encoder.Encode(motors)
  conn.Close()
}
