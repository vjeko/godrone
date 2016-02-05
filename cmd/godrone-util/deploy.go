package main

import (
	"fmt"
	"os"
  "os/exec"
  "io"
  "time"
  "net"
  "strings"

	"github.com/secsy/goftp"
)

const (
  firmwareBin = "firmware"
  killCmd     = "program.elf program.elf.respawner.sh " + firmwareBin
)

func compile() error {
	env := os.Environ()
	env = append(env, "GOOS=linux")
	env = append(env, "GOARCH=arm")

	binary := "go"
	args := []string{"build", "-o", "/tmp/firmware", "godrone/cmd/godrone"}
  cmd := exec.Command(binary, args...)
	cmd.Env = env
  return cmd.Run()
}

func transfer(dst_name string, src io.Reader) {

  var host = "192.168.1.1"
  var username = "root"
  var password = ""

	config := goftp.Config{
		User:               username,
		Password:           password,
		ConnectionsPerHost: 10,
		Timeout:            10 * time.Second,
		Logger:             nil,
	}

  fmt.Printf("Dialing into %s... ", host)
	client, err := goftp.DialConfig(config, host)
	if err != nil {
		panic(err)
	}
  fmt.Println("done.")

	// Upload a file from disk
  fmt.Printf("Writing to file %s... ", dst_name)
	err = client.Store(dst_name, src)
	if err != nil {
		panic(err)
	}
  fmt.Println("done.")
}

func read(conn *net.TCPConn, term string) *string {

  for {
    reply := make([]byte, 10240)

    _, err := conn.Read(reply)
    if err != nil {
        fmt.Println("Write to server failed:", err.Error())
        os.Exit(1)
    }

    str := string(reply)
    if strings.Contains(str, term) {
      return &str
    }
  }
}


func write(conn *net.TCPConn, cmd string) {
    cmd = fmt.Sprintf("%s\n", cmd)
    _, err := conn.Write([]byte(cmd))
    if err != nil {
        fmt.Println("Write to server failed:", err.Error())
        os.Exit(1)
    }
  }


func run(cmd string) bool {
  return runImpl(cmd, "#")
}

func runImpl(cmd string, condition string) bool {
    tcpAddr, err := net.ResolveTCPAddr("tcp4", "192.168.1.1:23")

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        fmt.Println("Dial failed:", err.Error())
        os.Exit(1)
    }

    read(conn, "#")
    write(conn, cmd)
    read(conn, condition)
    if strings.Compare(condition, "#") != 0 {
      return true
    }

    write(conn, "echo $?")
    result := read(conn, "#")
    *result = strings.TrimPrefix(*result, "echo $?\r\n")

    conn.Close()
    return strings.HasPrefix(*result, "0\r\n")
}

func main() {

  var err error

  cmd := "rm -rf /data/video/firmware"
  fmt.Printf("Executing command %s... ", cmd)
  run(cmd)
  fmt.Println("done.")

  fmt.Printf("Compiling the firmware... ")
  err = compile()
  if err != nil {
    fmt.Println("failed")
    os.Exit(-1)
  }
  fmt.Println("done.")

	bigFile, err := os.Open("/tmp/firmware")
	if err != nil {
		panic(err)
	}

  transfer("firmware", bigFile)

  cmd = "killall -9 " + killCmd
  fmt.Printf("Executing command %s... ", cmd)
  run(cmd)
  fmt.Println("done.")

  cmd = "chmod a+x /data/video/firmware"
  fmt.Printf("Executing command %s... ", cmd)
  if !run(cmd) {
    fmt.Println("failed.")
    os.Exit(-1)
  }

  cmd = "/data/video/firmware"
  fmt.Printf("Executing command %s... ", cmd)
  if !runImpl(cmd, "Up, up and away!") {
    fmt.Println("failed.")
    os.Exit(-1)
  }

  fmt.Println("done.")

}


