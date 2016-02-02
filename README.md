# Cross-compile:

http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5

export GOPATH=/Users/vjeko/dev/gocode
cd /Users/vjeko/dev/gocode/src
git clone git@github.com:vjeko/godrone.git
env GOOS=linux GOARCH=arm go build godrone/cmd/godrone

ftp root@192.168.1.1
put firmware

telnet 192.168.1.1
killall -9 program.elf.respawner.sh
killall -9 program.elf
cd /data/video/
./firmware

# godrone

GoDrone is a firmware for the Parrot AR Drone 2.0. It is developed
by Felix Geisendörfer and others using the Go programming language.

There is no affiliation with Parrot and running this firmware
may void your warranty.

Please read the docs at: http://www.godrone.io/en/latest/
