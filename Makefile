all: dagr

deps:
	go get -d -v
	go get github.com/GeertJohan/go.rice/rice

dagr-dev: *.go
	go build -o $(GOPATH)/bin/dagr-dev .

dagr: dagr-dev resources/templates/*.tmpl resources/static/*.js resources/static/*.css
	cp $(GOPATH)/bin/dagr-dev $(GOPATH)/bin/dagr
	$(GOPATH)/bin/rice append --exec $(GOPATH)/bin/dagr

clean:
	go clean
	rm -f dagr-dev

.PHONY: all clean deps
