dev:
	@go run main.go

build:
	@echo "Started building..."
	@go build -o bin/gotoleg
	@echo "Done."

buildlinux:
	@echo "Started building..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/gotoleg
	@echo "Done."