VERSION=$(shell git describe --tags)

build:
	go build -ldflags "-X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)"

install:
	go install -ldflags "-X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)"

release:
	gox --arch 'amd64 386' --os 'windows linux darwin' --output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" -ldflags "-s -w -X github.com/jrc2139/vimonade/lemon.Version=$(VERSION)"
	zip      pkg/vimonade_windows_386.zip     dist/vimonade_windows_386/vimonade.exe   -j
	zip      pkg/vimonade_windows_amd64.zip   dist/vimonade_windows_amd64/vimonade.exe -j
	tar zcvf pkg/vimonade_linux_386.tar.gz    -C dist/vimonade_linux_386/    vimonade
	tar zcvf pkg/vimonade_linux_amd64.tar.gz  -C dist/vimonade_linux_amd64/  vimonade
	tar zcvf pkg/vimonade_darwin_386.tar.gz   -C dist/vimonade_darwin_386/   vimonade
	tar zcvf pkg/vimonade_darwin_amd64.tar.gz -C dist/vimonade_darwin_amd64/ vimonade

clean:
	rm -rf dist/
	rm -f pkg/*.tar.gz pkg/*.zip
