NAME=awsping
EXEC=${NAME}
GOVER=1.7.1
ENVNAME=${NAME}${GOVER}

build: deps
	go build -o ${EXEC} main.go

deps:
	echo "no deps yet"

run:
	./${EXEC}

test:
	@go test -v

release:
	git tag `grep "version" main.go | grep -o -E '[0-9]\.[0-9]\.[0-9]{1,2}'`
	git push --tags origin master


#
# For virtual environment create with
# https://github.com/ekalinin/envirius
#
env-create: env-init env-deps

env-init:
	@bash -c ". ~/.envirius/nv && nv mk ${ENVNAME} --go-prebuilt=${GOVER}"

env-build:
	@bash -c ". ~/.envirius/nv && nv do ${ENVNAME} 'make build'"

env-deps:
	@bash -c ". ~/.envirius/nv && nv do ${ENVNAME} 'make deps'"

env:
	@bash -c ". ~/.envirius/nv && nv use ${ENVNAME}"

