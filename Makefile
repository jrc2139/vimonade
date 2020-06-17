VERSION=$(shell git describe --tags)

cert: ## Create certificates to encrypt the gRPC connection
	openssl genrsa -out ca.key 4096
	openssl req -new -x509 -key ca.key -sha256 -subj "/C=US/ST=NJ/O=CA, Inc." -days 365 -out ca.cert
	openssl genrsa -out certs/service.key 4096
	openssl req -new -key certs/service.key -out service.csr -config certificate.conf
	openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial \
		-out certs/service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext

build:
	rice embed-go
	go build -ldflags "-X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)" 

install:
	go install -ldflags "-X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)"

release:
	gox --arch 'amd64 386' --os 'windows linux darwin' --output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" -ldflags "-s -w -X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)"
	zip      release_pkg/vimonade_windows_386.zip     dist/vimonade_windows_386/vimonade.exe   -j
	zip      release_pkg/vimonade_windows_amd64.zip   dist/vimonade_windows_amd64/vimonade.exe -j
	tar zcvf release_pkg/vimonade_linux_386.tar.gz    -C dist/vimonade_linux_386/    vimonade
	tar zcvf release_pkg/vimonade_linux_amd64.tar.gz  -C dist/vimonade_linux_amd64/  vimonade
	tar zcvf release_pkg/vimonade_darwin_386.tar.gz   -C dist/vimonade_darwin_386/   vimonade
	tar zcvf release_pkg/vimonade_darwin_amd64.tar.gz -C dist/vimonade_darwin_amd64/ vimonade

clean:
	rm -rf dist/
	rm -f release_pkg/*.tar.gz release_pkg/*.zip
