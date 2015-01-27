version=`git rev-parse --short HEAD`

all: dagr

deps:
	go get -d -v

dagr-dev: *.go
	go build -ldflags "-X main.Revision $(version)" -o $(GOPATH)/bin/dagr-dev .

dagr: dagr-dev
	cp $(GOPATH)/bin/dagr-dev $(GOPATH)/bin/dagr

clean:
	go clean
	rm -f bin/dagr*
	rm -rf src/

.PHONY: all clean deps
