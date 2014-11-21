all: dagr
	
.PHONY: deps

deps:
	go get

dagr-dev: *.go
	go build -o dagr-dev .

dagr: dagr-dev resources/index.html.tmpl resources/info.html.tmpl resources/dagr.css
	cp dagr-dev dagr && nrsc dagr ./resources

clean:
	go clean
	rm -f dagr-dev

.PHONY: all clean

