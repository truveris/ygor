all: ygord/ygord ygor-minion/ygor-minion

test: ygord/ygord ygor-minion/ygor-minion
	cd ygord && make test
	cd ygor-minion && make test

ygord/ygord:
	cd ygord && make

ygor-minion/ygor-minion:
	cd ygor-minion && make

fmt:
	go fmt
	cd ygord && go fmt
	cd ygor-minion && go fmt

clean:
	cd ygord && make clean
	cd ygor-minion && make clean
