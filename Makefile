all: ygor-body/ygor-body ygor-truveris/ygor-truveris ygor-minion/ygor-minion

ygor-body/ygor-body:
	cd ygor-body && make

ygor-truveris/ygor-truveris:
	cd ygor-truveris && make

ygor-minion/ygor-minion:
	cd ygor-minion && make

clean:
	cd ygor-body && make clean
	cd ygor-truveris && make clean
	cd ygor-minion && make clean
