all: b

assets:
	go-assets-builder views >src/assets.go

b:
	mkdir -p tmp
	go-assets-builder views >src/assets.go
	go build -o tmp/otpbase src/server.go src/otp.go src/assets.go
