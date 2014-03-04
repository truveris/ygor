all: ygor/ygor ygor-minion/ygor-minion

ygor/ygor:
	cd ygor && make

ygor-minion/ygor-minion:
	cd ygor-minion && make

test: ygor/ygor ygor-minion/ygor-minion
	cd tests/ && make test

clean:
	cd ygor && make clean
	cd ygor-minion && make clean
