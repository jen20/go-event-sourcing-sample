all:
	@go generate && go build

clean:
	@rm -f goes
	@rm -f *_string.go
