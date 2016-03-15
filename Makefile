default: build.test

clean:
	rm -rf build

check:
	golint ./... | egrep -v '^vendor/'
	go vet ./... 2>&1 | egrep -v '^(vendor/|exit status 1)'
	unused ./... | egrep -v '^vendor/'

build.test:
	go test -v ./...

prepare:
	mkdir -p build/linux
	mkdir -p build/osx

build.local: prepare
	godep go build -o build/chimp-server -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation  ./cmd/chimp-server
	godep go build -o build/chimp -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation  ./cmd/chimp

build.linux: prepare
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o build/linux/chimp-server -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation ./cmd/chimp-server
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o build/linux/chimp -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation ./cmd/chimp


build.osx: prepare
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 godep go build -o build/osx/chimp-server -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation ./cmd/chimp-server
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 godep go build -o build/osx/chimp -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation ./cmd/chimp

dev.install:
	godep go install -ldflags "-X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`" -tags zalandoValidation github.com/zalando-techmonkeys/chimp/...
