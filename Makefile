default: build.local

clean:
	rm -rf build

prepare: clean
	mkdir -p build/linux
	mkdir -p build/osx

build.local: prepare
	go build -o build/chimp-server -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation  ./cmd/chimp-server
	go build -o build/chimp -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation  ./cmd/chimp

build.linux: prepare
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o build/linux/chimp-server -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation ./cmd/chimp-server
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o build/linux/chimp -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation ./cmd/chimp


build.osx: prepare
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 godep go build -o build/osx/chimp-server -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation ./cmd/chimp-server
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 godep go build -o build/osx/chimp -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation ./cmd/chimp

