package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/felixge/godrone/attitude"
	"github.com/felixge/godrone/control"
	"github.com/felixge/godrone/drivers/motorboard"
	"github.com/felixge/godrone/drivers/navboard"
	"github.com/felixge/godrone/http"
	"github.com/felixge/log"
	gohttp "net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var c = flag.String("c", "", "Absolute or relative path to config file.")

type Config struct {
	NavboardTTY   string
	MotorboardTTY string
	RollPID       []float64
	PitchPID      []float64
	YawPID        []float64
	HttpAddr      string
}

var (
	defaultRollPitchPID = []float64{0.04, 0, 0.002}

	green  = motorboard.Leds(motorboard.LedGreen)
	orange = motorboard.Leds(motorboard.LedOrange)
	red    = motorboard.Leds(motorboard.LedRed)
)

var DefaultConfig = Config{
	NavboardTTY:   "/dev/ttyO1",
	MotorboardTTY: "/dev/ttyO0",
	RollPID:       defaultRollPitchPID,
	PitchPID:      defaultRollPitchPID,
	YawPID:        []float64{0.04, 0, 0}, // disabled, needs magnotometer to work well
	HttpAddr:      ":80",
}

type Instances struct {
	log        *log.Logger
	navboard   *navboard.Navboard
	motorboard *motorboard.Motorboard
	attitude   *attitude.Complementary
	control    *control.Control
	http       *http.Handler
}

func main() {
	flag.Parse()

	config := DefaultConfig
	if *c != "" {
		if err := LoadConfig(*c, &config); err != nil {
			panic(err)
		}
	}
	i, err := NewInstances(config)
	if err != nil {
		panic(err)
	}
	i.log.Info("Starting godrone")
	defer i.motorboard.Close()

	i.motorboard.SetLeds(green)
	time.Sleep(500 * time.Millisecond)
	i.motorboard.SetLeds(red)

	i.log.Info("Calibrating sensors")
	for {
		if err := i.navboard.Calibrate(); err == nil {
			break
		}
	}
	i.motorboard.SetLeds(green)

	navDataCh := make(chan navboard.Data)
	go readNavData(i.navboard, navDataCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)

	go gohttp.ListenAndServe(config.HttpAddr, i.http)

	i.log.Info("Entering main loop")
mainLoop:
	for {
		select {
		case navData := <-navDataCh:
			attitudeData := i.attitude.Update(navData.Data)
			motorSpeeds := i.control.Update(attitudeData)
			if err := i.motorboard.SetSpeeds(motorSpeeds); err != nil {
				i.log.Error("Could not set motor speeds. err=%s", err)
			}
			i.http.Update(navData, attitudeData)
		case sig := <-sigCh:
			i.log.Info("Received signal=%s, shutting down", sig)
			break mainLoop
		}
	}
}

func readNavData(board *navboard.Navboard, ch chan<- navboard.Data) {
	for {
		navData, err := board.NextData()
		if err != nil {
			continue
		}
		select {
		case ch <- navData:
		default:
		}
	}
}

func NewInstances(c Config) (i Instances, err error) {
	i.log = log.DefaultLogger
	i.navboard = navboard.NewNavboard(c.NavboardTTY, i.log)
	i.motorboard, err = motorboard.NewMotorboard(c.MotorboardTTY)
	if err != nil {
		return
	}
	i.attitude = attitude.NewComplementary()
	i.control = control.NewControl(c.RollPID, c.PitchPID, c.YawPID)
	i.http = http.NewHandler(http.Config{
		Control: i.control,
		Log:     i.log,
	})
	return
}

func LoadConfig(file string, config *Config) error {
	if string(file[0]) != "/" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		file = filepath.Join(wd, file)
	}
	_, err := toml.DecodeFile(file, &config)
	return err
}
