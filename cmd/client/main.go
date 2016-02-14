package main

import (

  //"./joystick"
  "godrone/cmd/client/joystick"
  "net"
  "os"
  "fmt"
  "log"
  //"math"

  "encoding/gob"
  "godrone"

	"github.com/nsf/termbox-go"
  "github.com/veandco/go-sdl2/sdl"
)

func printRow(row int, str string) {
  for i := 0; i < len(str); i++ {
    termbox.SetCell(i, row, rune(str[i]), 
    termbox.ColorWhite, termbox.ColorBlack)
  }
}

func keyEvents() {

  for {
    event := termbox.PollEvent()
    if(event.Key == 27) {
      os.Exit(0)
    }
  }
}

func pollJoystick(c chan [5]int16, m chan godrone.MotorPWM) {
  js, err := joystick.GetJs()
  if err != nil {
    panic(err)
  }

  for {
    axes := joystick.GetAxes(js)
    sdl.Delay(100)
    c <- axes

    var speed float64 = 0.0
    var oper int64 = 0
    oper += 32768 - int64(axes[2])
    if (axes[2] != 0) {
      speed = float64(oper)/65536
    }

    motors := godrone.MotorPWM{[4]float64{speed, speed, speed, speed}}
    m <- motors

    printRow(1, fmt.Sprintf("%v          ", speed))
    printRow(2, fmt.Sprintf("%v          ", motors))
  }
}

func printScreen(c chan [5]int16) {

  err := termbox.Init()
  if err != nil {
          panic(err)
  }
  defer termbox.Close()

  for {
    select {
    case axes := <-c:
      printRow(0, fmt.Sprintf("%v", axes))
      termbox.Flush()
    }
  }
}

func ctrlMotors(c chan godrone.MotorPWM) {
  conn, err := net.Dial("tcp", "192.168.1.1:666")
  if err != nil {
    panic(err)
  }
  defer conn.Close()


  encoder := gob.NewEncoder(conn)
  for {
    select {
    case motors := <-c:

      printRow(3, fmt.Sprintf("Sending %v", motors))
      val := godrone.MotorPWM(motors)
      encoder.Encode(val)
    }
  }
}

func main() {
  log.Println("Starting the client.")
  axesCh := make(chan [5]int16)
  motorCh := make(chan godrone.MotorPWM)

  go ctrlMotors(motorCh)
  go keyEvents()
  go printScreen(axesCh)

  pollJoystick(axesCh, motorCh)
}
