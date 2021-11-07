NAME=awsping
EXEC=${NAME}
BUILD_DIR=build
BUILD_OS="windows darwin freebsd linux"
BUILD_ARCH="amd64 386"
BUILD_DIR=build
SRC_CMD=cmd/awsping/main.go

build:
	go build -race -o ${EXEC} ${SRC_CMD}

# make run ARGS="-h"
run:
	go run cmd/awsping/main.go $(ARGS)

test:
	@go test -cover .

release: buildall
	git tag `grep "Version" utils.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	git push --tags origin master

clean:
	@rm -f ${EXEC}
	@rm -f ${BUILD_DIR}/*
	@go clean

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

docker:
	docker build -t awsping .

docker-run: docker
	docker run awsping -verbose 2 -repeats 2
