all: ygord/ygord ygorlet/ygorlet

test: ygord/ygord ygorlet/ygorlet
	cd ygord && make test
	cd ygorlet && make test

ygord/ygord:
	cd ygord && make

ygorlet/ygorlet:
	cd ygorlet && make

fmt:
	go fmt
	cd ygord && go fmt
	cd ygorlet && go fmt

clean:
	cd ygord && make clean
	cd ygorlet && make clean
