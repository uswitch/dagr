version=`git rev-parse --short HEAD`

all: dagr

deps:
	go get -d -v
	go get github.com/GeertJohan/go.rice/rice

dagr-dev: *.go
	go build -ldflags "-X main.Revision $(version)" -o $(GOPATH)/bin/dagr-dev .

dagr: dagr-dev resources/templates/*.tmpl resources/static/*.js resources/static/*.css
	cp $(GOPATH)/bin/dagr-dev $(GOPATH)/bin/dagr
	$(GOPATH)/bin/rice append --exec $(GOPATH)/bin/dagr

clean:
	go clean
	rm -f bin/dagr*
	rm -rf src/

.PHONY: all clean deps
