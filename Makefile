NAME=upbot
VERSION = $(shell git describe --always --tags)
BUILD_FLAGS = -mod vendor
all:
	go build $(BUILD_FLAGS)
build:
	CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS)
	docker build -t $(NAME):$(VERSION) -t $(NAME):latest .
up:
	docker-compose up -d
down:
	docker-compose down
clean: down
	rm -rf $(NAME)
	docker rmi -f $(shell docker images -f "dangling=true" -q) 2> /dev/null; true
	docker rmi -f $(NAME):latest $(NAME):$(VERSION) 2> /dev/null; true
