package joystick

import (
  "github.com/veandco/go-sdl2/sdl"
  "log"
  "errors"
)


func GetJs() (*sdl.Joystick, error) {
  sdl.Init(sdl.INIT_JOYSTICK)
  numJs := sdl.NumJoysticks()
  if (numJs == 0) {
    return nil, errors.New("Joystick missing.")
  }

  jsIdx := -1
  for i := 0; i < numJs; i++ {
    if sdl.JoystickNameForIndex(i) == "T.Flight Hotas X" {
      jsIdx = i
      break
    }
  }

  if jsIdx == -1 {
    return nil, errors.New("Unsuported joystick!")
  }

  return sdl.JoystickOpen(jsIdx), nil
}

func GetAxes(js* sdl.Joystick) [5]int16 {
  var axes [5]int16

    sdl.Update()
    for i := 0; i < js.NumAxes(); i++ {
      axes[i] = js.GetAxis(i)
    }

    return axes
}


func main() {
  js, err := GetJs()
  if err != nil {
    panic(err)
  }

  for {
    axes := GetAxes(js)
    sdl.Delay(100)
    log.Println(axes)
  }
}
