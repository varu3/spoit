GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
export GO111MODULE := on

.PHONY: test binary install clean dist

cmd/spoit/spoit: *.go cmd/spoit/*.go
	cd cmd/spoit && go build -ldflags "-s -w -X main.Version=${GIT_VER}" -gcflags="-trimpath=${PWD}"

install: cmd/spoit/spoit
	install cmd/spoit/spoit ${GOPATH}/bin

test:
	go test -race .
	go test -race ./cmd/spoit

clean:
	rm -f cmd/spoit/spoit
	rm -rf dist/

dist:
	CGO_ENABLED=0 \
		goxz -pv=$(GIT_VER) \
		-build-ldflags="-s -w -X main.Version=${GIT_VER}" \
		-os=darwin,linux -arch=amd64 -d=dist ./cmd/spoit

release:
	ghr -u varusan -r spoit -n "$(GIT_VER)" $(GIT_VER) dist/