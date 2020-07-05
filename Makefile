VERSION=$(shell git describe --tags)
INSECURE_DIR="./cmd/vimonade"
SECURE_DIR="./cmd/secure"
BIN="vimonade"

cert: ## Create certificates to encrypt the gRPC connection
	openssl genrsa -out ca.key 4096
	openssl req -new -x509 -key ca.key -sha256 -subj "/C=US/ST=NJ/O=CA, Inc." -days 3650 -out ca.cert
	openssl genrsa -out certs/service.key 4096
	openssl req -new -key certs/service.key -out service.csr -config certificate.conf
	openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial \
		-out certs/service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext

build:
	go build -o ${BIN} -v -ldflags "-X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)" ${INSECURE_DIR}

install:
	go install -v -ldflags "-X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)" ${INSECURE_DIR}

install-remote:
	make build-insecure
	scp ${BIN} $1 

build-secure:
	 cd ${SECURE_DIR} && rice embed-go
	 go build -o ${BIN} -v -ldflags "-X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)" ${SECURE_DIR}

install-secure:
	# make cert
	make build-secure
	mv ${BIN} ${GOPATH}/bin

install-secure-remote:
	make build-secure
	./install-remote.sh ${BIN}

release:
	mkdir ${INSECURE_DIR}/dist && mkdir ${INSECURE_DIR}/pkg
	cd ${INSECURE_DIR} && gox --arch 'amd64 386' --os 'windows linux darwin' --output "../../dist/${BIN}_{{.OS}}_{{.Arch}}/${BIN}" -ldflags "-s -w -X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)"
	zip      pkg/${BIN}_windows_386.zip     dist/${BIN}_windows_386/${BIN}.exe   -j
	zip      pkg/${BIN}_windows_amd64.zip   dist/${BIN}_windows_amd64/${BIN}.exe -j
	tar zcvf pkg/${BIN}_linux_386.tar.gz    -C dist/${BIN}_linux_386/ 		 ${BIN}
	tar zcvf pkg/${BIN}_linux_amd64.tar.gz  -C dist/${BIN}_linux_amd64/ 	 ${BIN}
	tar zcvf pkg/${BIN}_darwin_386.tar.gz   -C dist/${BIN}_darwin_386/ 		 ${BIN}
	tar zcvf pkg/${BIN}_darwin_amd64.tar.gz -C dist/${BIN}_darwin_amd64/ 	 ${BIN}

clean:
	rm -rf dist/
	rm -f pkg/*.tar.gz pkg/*.zip
	rm -rf ${INSECURE_DIR}/dist/
	rm -rf ${INSECURE_DIR}/pkg/
