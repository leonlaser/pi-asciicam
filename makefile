all: build-linux build-arm6 build-arm7 build-macos

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux-amd64/asciicam-imux ./asciicam-imux/
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux-amd64/asciicam-server ./asciicam-server/

build-arm6:
	GOOS=linux GOARCH=arm GOARM=6 go build -o ./bin/linux-arm6/asciicam-imux ./asciicam-imux/
	GOOS=linux GOARCH=arm GOARM=6 go build -o ./bin/linux-arm6/asciicam-server ./asciicam-server/

build-arm7:
	GOOS=linux GOARCH=arm GOARM=7 go build -o ./bin/linux-arm7/asciicam-imux ./asciicam-imux/
	GOOS=linux GOARCH=arm GOARM=7 go build -o ./bin/linux-arm7/asciicam-server ./asciicam-server/

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/macos-amd64/asciicam-imux ./asciicam-imux/
	GOOS=darwin GOARCH=amd64 go build -o ./bin/macos-amd64/asciicam-server ./asciicam-server/

docker-image:
	cp ./bin/linux-amd64/asciicam-imux ./asciicam-imux/
	docker build -t leonlaser/pi-asciicam-imux -f ./asciicam-imux/Dockerfile ./asciicam-imux
	rm ./asciicam-imux/asciicam-imux

docker-image-arm:
	cp ./bin/linux-arm7/asciicam-imux ./asciicam-imux/
	docker build -t leonlaser/pi-asciicam-imux-arm -f ./asciicam-imux/Dockerfile.arm ./asciicam-imux
	rm ./asciicam-imux/asciicam-imux

.PHONY: build-linux build-arm6 build-arm7 build-macos
