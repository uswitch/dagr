all: dagr

deps:
	go get
	go get github.com/GeertJohan/go.rice
	go get github.com/GeertJohan/go.rice/rice

dagr-dev: *.go
	go build -o dagr-dev .

dagr: dagr-dev resources/*.tmpl resources/*.js resources/*.css
	cp dagr-dev dagr && rice append --exec dagr

clean:
	go clean
	rm -f dagr-dev

.PHONY: all clean deps
