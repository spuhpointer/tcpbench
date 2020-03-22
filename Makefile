
export GOPATH=$(shell pwd)

all:
	make -C src

clean:
	go get -u github.com/endurox-dev/endurox-go
	make -C src clean


.PHONEY: clean all
