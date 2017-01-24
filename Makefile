GO := $(shell command -v go 2> /dev/null)

default: bin

env:
ifndef GO
	$(error "go was not found.  Please install go and setup the go env)
endif

bin/stardog-graviton: env
	./scripts/build-local.sh

bin: bin/stardog-graviton

test: bin/stardog-graviton
	go test -v -cover github.com/stardog-union/stardog-graviton/integration github.com/stardog-union/stardog-graviton/aws github.com/stardog-union/stardog-graviton/sdutils github.com/stardog-union/stardog-graviton

clean:
	rm -f aws/data.go
	rm -f data.go
	rm -f ${GOPATH}/bin/stardog-graviton
	rm -f etc/version

.PHONY: bin clean test
