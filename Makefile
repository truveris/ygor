all: ygor/ygor ygor-truveris/ygor-truveris ygor-minion/ygor-minion

ygor/ygor:
	cd ygor && make

ygor-truveris/ygor-truveris:
	cd ygor-truveris && make

ygor-minion/ygor-minion:
	cd ygor-minion && make

test:
	cd tests/ && make test

clean:
	cd ygor && make clean
	cd ygor-truveris && make clean
	cd ygor-minion && make clean
