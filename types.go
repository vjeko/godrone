package godrone

type MotorPWM struct {
    Speed  float64
}

type Pong struct {
    Nav    Navdata
}

const (
  MAX_MOTOR = 0.7
)
