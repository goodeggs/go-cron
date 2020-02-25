.PHONY: dist

dist:
	@mkdir -p dist/
	@docker build -t goodeggs/go-cron:latest .
	@docker run --entrypoint '' -e CGO_ENABLED=0 -e GOOS=linux -v "$$PWD/dist:/dist" goodeggs/go-cron:latest go build -ldflags '-w' -o /dist/go-cron-linux-amd64 -a ./...
