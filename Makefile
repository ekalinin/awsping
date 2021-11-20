NAME=awsping
EXEC=${NAME}
BUILD_DIR=build
BUILD_OS="windows darwin freebsd linux"
BUILD_ARCH="amd64 386"
BUILD_DIR=build
SRC_CMD=cmd/awsping/main.go
VERSION=`grep "Version" utils.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`

build:
	go build -race -o ${EXEC} ${SRC_CMD}

clean:
	@rm -f ${EXEC}
	@rm -f ${BUILD_DIR}/*
	@go clean

#
# Tests, linters
#

lint:
	golint

# make run ARGS="-h"
run:
	go run cmd/awsping/main.go $(ARGS)

test: lint
	@go test -cover .

#
# Release
#
check-version:
ifdef VERSION
	@echo Current version: $(VERSION)
else
	$(error VERSION is not set)
endif

check-master:
ifneq ($(shell git rev-parse --abbrev-ref HEAD),master)
	$(error You're not on the "master" branch)
endif

release: check-master check-version
	git tag v${VERSION} && \
	git push origin v${VERSION}

release-test: check-master check-version
	goreleaser release --snapshot --rm-dist

buildall: clean
	@mkdir -p ${BUILD_DIR}
	@for os in "${BUILD_OS}" ; do \
		for arch in "${BUILD_ARCH}" ; do \
			echo " * build $$os for $$arch"; \
			GOOS=$$os GOARCH=$$arch go build -ldflags "-s" -o ${BUILD_DIR}/${EXEC} ${SRC_CMD}; \
			cd ${BUILD_DIR}; \
			tar czf ${EXEC}.$$os.$$arch.tgz ${EXEC}; \
			cd - ; \
		done done
	@rm ${BUILD_DIR}/${EXEC}

#
# Docker
#

docker:
	docker build -t awsping .

docker-run: docker
	docker run awsping -verbose 2 -repeats 2
