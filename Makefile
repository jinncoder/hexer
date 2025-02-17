VERSION=$(shell grep "VERSION = " cli/cmd/root.go | cut -d"\"" -f2 | tr -d '\n')
BUILD=$(shell git rev-parse HEAD)
BASEDIR=./dist
DIR=${BASEDIR}/temp

LDFLAGS=-ldflags "-s -w -X main.build=${BUILD} -buildid=${BUILD}"
GCFLAGS=-gcflags=all=-trimpath=$(shell pwd)
ASMFLAGS=-asmflags=all=-trimpath=$(shell pwd)

GOFILES=`go list -buildvcs=false ./...`
GOFILESNOTEST=`go list -buildvcs=false ./... | grep -v test`

# Make Directory to store executables
$(shell mkdir -p ${DIR})

all: tidy linux freebsd edr docs
# goreleaser build --config .goreleaser.yml --rm-dist --skip-validate

edr:
	sqlite3 storage/data.db -init mkdocs/sqlite-schema-diagram.sql "" > mkdocs/schema.dot
	dot -Tsvg mkdocs/schema.dot > mkdocs/schema.svg

freebsd: lint security docs
	@env CGO_ENABLED=1 GOOS=freebsd GOARCH=amd64 go build -trimpath ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} -o ${DIR}/hexer-freebsd_amd64 main.go

linux: lint security docs
	@env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} -o ${DIR}/hexer-linux_amd64 main.go

docs:
	rm -rf mkdocs/docs/usage/ || echo ''
	@mkdir -p mkdocs/docs/usage/
	@go run main.go doc
	@mv docs/* mkdocs/docs/usage/

tidy:
	@go mod tidy

update: tidy
	@go get -v ./...
	@go get -u all

dep: ## Get the dependencies # go get github.com/goreleaser/goreleaser
	@git config --global url."git@github.com:".insteadOf "https://github.com/"
	@go install github.com/boumenot/gocover-cobertura@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest

lint: ## Lint the files
	@env CGO_ENABLED=1 go fmt ${GOFILES}
	@env CGO_ENABLED=1 go vet ${GOFILESNOTEST}

security: dep tidy
	@go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...
	# TODO @go run github.com/go-critic/go-critic/cmd/gocritic@latest check -enableAll -disable='#experimental,#opinionated' ./...
	# TODO @go run github.com/google/osv-scanner/cmd/osv-scanner@latest -r . || echo "oh snap!"

release:
	@goreleaser release --config .github/goreleaser.yml

clean:
	@rm -rf ${BASEDIR}
	@rm -rf mkdocs/docs/usage/

.PHONY: all freebsd linux docs submodule tidy update dep lint security test release clean
