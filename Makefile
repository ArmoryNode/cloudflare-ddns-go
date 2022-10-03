BINARY_NAME=cloudflare-ddns

build:
	go build -o ${BINARY_NAME}.exe main.go

clean:
	go clean
	rm ${BINARY_NAME}.exe