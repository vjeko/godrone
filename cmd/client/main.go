package main

import (
  "godrone/cmd/client/joystick"
  "net"
  "os"
  "fmt"
  "log"

  "encoding/gob"
  "godrone"

	"github.com/nsf/termbox-go"
  "github.com/veandco/go-sdl2/sdl"
)

const (
  THROTTLE_AXIS = 2
  TRIM_AXIS     = 4
)

const (
  POLL_DELAY  = 100
  TRIM_CHANGE = POLL_DELAY*4
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

func pollJoystick(c chan [5]int16, m chan *godrone.MotorPWM) {
  js, err := joystick.GetJs()
  if err != nil {
    panic(err)
  }

  var trim float64 = 0.1
  for {
    axes := joystick.GetAxes(js)
    sdl.Delay(100)
    c <- axes


    if axes[TRIM_AXIS] != 0 {
      var oper int64 = int64(axes[TRIM_AXIS]) - 128
      delta :=  (float64(oper)/65536)/TRIM_CHANGE
      trim += delta
    }

    if trim < 0 {
      trim = 0.0
    }

    if trim > godrone.MAX_MOTOR {
      trim = godrone.MAX_MOTOR
    }

    var throttle float64 = 0.0
    if axes[THROTTLE_AXIS] != 0 {
      var oper int64 = 32768 - int64(axes[THROTTLE_AXIS])
      throttle = (float64(oper)/65536) * trim
    }


    motors := godrone.MotorPWM{throttle}
    m <- &motors

    printRow(1, fmt.Sprintf("THrottle: %f          ", throttle))
    printRow(2, fmt.Sprintf("Trim:     %f          ", trim))
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

func handleNav() {
  ln, err := net.Listen("tcp", ":7777")
  if err != nil {
    log.Println("Error: ", err)
    return
  }

  for {
    conn, err := ln.Accept()
    if err != nil {
      continue
    }

    handleNavImpl(conn)
  }
}

func handleNavImpl(conn net.Conn) {

  dec := gob.NewDecoder(conn)
  for {
    place := &godrone.Placement{}
    err := dec.Decode(place)
    if err != nil {
      log.Println(err)
      return
    }
    printRow(4, fmt.Sprintf("Sensor: %v %v %v         ",
      place.Roll,
      place.Pitch, place.Yaw))
  }
}

func ctrlMotors(c chan *godrone.MotorPWM) {
  conn, err := net.Dial("tcp", "192.168.1.1:666")
  if err != nil {
    panic(err)
  }
  defer conn.Close()

  encoder := gob.NewEncoder(conn)
  for {
    select {
    case packet := <-c:
      encoder.Encode(packet)
    }
  }
}

func main() {
  log.Println("Starting the client.")
  axesCh := make(chan [5]int16)
  motorCh := make(chan *godrone.MotorPWM)

	go handleNav()
  go ctrlMotors(motorCh)
  go keyEvents()
  go printScreen(axesCh)

  pollJoystick(axesCh, motorCh)
}
