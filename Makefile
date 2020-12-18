all:
	@go generate && go fmt ./... && go build

clean:
	@rm -f goes
	@rm -f *_string.go
