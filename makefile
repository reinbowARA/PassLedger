APP_NAME = passwordbook
BUILD_DIR = build

all: build

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) main.go

run:
	go run main.go

clean:
	rm -rf $(BUILD_DIR)

build-win:
	go build -ldflags="-H=windowsgui" -o $(BUILD_DIR)/$(APP_NAME).exe main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME) main.go

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME) main.go
