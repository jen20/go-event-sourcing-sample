all:
	@go generate && go build . && ./go-event-sourcing-sample

clean:
	@rm -f go-event-sourcing-sample

install-stringer:
	@go install golang.org/x/tools/cmd/stringer
