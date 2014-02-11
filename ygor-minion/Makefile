# This Makefile is mostly useful for raspberrypi deployments.
#

PI_SSH_HOST=pi@10.11.12.51

all:
	@echo nope

ygor-minion: main.go
	go build

clean:
	rm -f ygor-minion

piinstall:
	GOARCH=arm GOARM=5 GOOS=linux go build
	rsync -avz ygor-minion tunes ${PI_SSH_HOST}:ygor-minion/
	ssh ${PI_SSH_HOST} "pkill ygor-minion"
