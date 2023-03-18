dev:
	@go run main.go

build:
	@echo "Started building..."
	@go build -o bin/cash
	@echo "Done."

buildlinux:
	@echo "Started building..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/gocash
	@echo "Done."