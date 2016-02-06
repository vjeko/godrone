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
  host     = "192.168.1.1"
  telnet   = "23"
  username = "root"
  password = ""
)
const (
  bufferSize  = 10240
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

func transfer(dst_name string, src io.Reader) error {

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
    return err
	}
  fmt.Println("done.")

  fmt.Printf("Writing to file %s... ", dst_name)
	err = client.Store(dst_name, src)
	if err != nil {
    return err
	}
  fmt.Println("done.")
  return nil
}

func read(conn *net.TCPConn, term string) *string {

  for {
    reply := make([]byte, bufferSize)

    _, err := conn.Read(reply)
    if err != nil {
        panic(err)
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
        panic(err)
    }
  }


func Exec(cmd string) bool {
  return ExecCondition(cmd, "#")
}

func ExecCondition(cmd string, condition string) bool {
    tcpAddr, err := net.ResolveTCPAddr("tcp4", host + ":" + telnet)
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
  Exec(cmd)
  fmt.Println("done.")

  fmt.Printf("Compiling the firmware... ")
  err = compile()
  if err != nil {
    panic(err)
  }
  fmt.Println("done.")

	bigFile, err := os.Open("/tmp/firmware")
	if err != nil {
		panic(err)
	}

  err = transfer("firmware", bigFile)
  if err != nil {
		panic(err)
  }

  cmd = "killall -9 " + killCmd
  fmt.Printf("Executing command %s... ", cmd)
  _ = Exec(cmd)
  fmt.Println("done.")

  cmd = "chmod a+x /data/video/firmware"
  fmt.Printf("Executing command %s... ", cmd)
  if !Exec(cmd) {
    panic("failed.")
  }
  fmt.Println("done.")

  cmd = "/data/video/firmware"
  fmt.Printf("Executing command %s... ", cmd)
  if !ExecCondition(cmd, "Success.") {
    panic("failed.")
  }
  fmt.Println("done.")

}


