all: dagr

dagr-dev: main.go web/web.go program/program.go execute/execute.go git/git.go
	go build -o dagr-dev .

dagr: dagr-dev resources/index.html.tmpl
	cp dagr-dev dagr && nrsc dagr ./resources

clean:
	rm -rf dagr dagr-dev