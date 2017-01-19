version=`git rev-parse --short HEAD`

all: dagr ui.tgz

deps:
	go get -d -v

dagr-dev: *.go
	go build -ldflags "-X main.Revision=$(version)" -o $(GOPATH)/bin/dagr-dev .

dagr: dagr-dev
	cp $(GOPATH)/bin/dagr-dev $(GOPATH)/bin/dagr

ui.tgz: ui
	tar zcf ui.tgz ui

clean:
	go clean
	rm -f bin/dagr*
	rm -rf src/
	rm -f ui.tgz

.PHONY: all clean deps
