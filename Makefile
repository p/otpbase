all: b

b:
	mkdir -p tmp
	go-assets-builder views >src/assets.go
	go build -o tmp/otpbase src/server.go src/otp.go src/assets.go
