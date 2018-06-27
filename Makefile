all: b

b:
	mkdir -p tmp
	go build -o tmp/totpbase src/server.go src/otp.go
