all: b

assets:
	go-assets-builder views >src/assets.go

b:
	mkdir -p tmp
	go-assets-builder views >src/assets.go
	go build -o tmp/otpbase src/server.go src/otp.go src/assets.go src/sms.go

fmt:
	for f in src/*.go; do go fmt $$f && sed -i -e 's/	/  /g' $$f; done

# https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/
docker:
	mkdir -p tmp
	go-assets-builder views >src/assets.go
	CGO_ENABLED=0 GOOS=linux go build -o tmp/otpbase.docker src/server.go src/otp.go src/assets.go src/sms.go
	docker build -t otpbase .
