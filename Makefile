run:
	@go run main.go

build:
	@go build -o jfrog-cleaner

build-linux:
	@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build  -o jfrog-cleaner