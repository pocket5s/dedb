.PHONY: all test clean mocks setup run cover genproto
GOCMD=go

test:
	@if [ ! -f /go/bin/gotest ]; then \
		echo "installing gotest..."; \
		go get github.com/rakyll/gotest; \
		go install github.com/rakyll/gotest; \
	fi 
	gotest -v ./internal

cover:
	gotest -covermode=count -coverpkg=. -coverprofile=profile.cov ./... fmt
	go tool cover -func=profile.cov
	go tool cover -html=profile.cov -o=coverage.html

clean:
	go clean
	if test -f go.mod; then echo ""; else rm app; fi
	rm -fr account_domain events tenant_domain

genproto:
	go get google.golang.org/protobuf/encoding/protojson@v1.26.0
	go install github.com/golang/protobuf/protoc-gen-go
	rm -fr ./*.pg.go
	protoc --go_out=plugins=grpc:. \
      --proto_path=./ \
      dedb.proto
	mv github.com/pocket5s/dedb/* .
	rm -fr github.com


build:
	@-$(MAKE) -s clean
	#@-$(MAKE) -s genproto
	go build -o app cmd/main.go
	chmod +x app
