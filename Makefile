GITTAG := `git describe --tags`
VERSION := `git describe --abbrev=0 --tags`
RELEASE := `git rev-list $(shell git describe --abbrev=0 --tags).. --count`
DESTDIR?=dist
BUILD_TIME := `date +%FT%T%z`
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS := -ldflags "-X main.GitTag=${GITTAG} -X main.Version=${VERSION} -X main.Build=${BUILD_TIME}"

.PHONY: binary
binary: clean build

.PHONY: clean
clean:
	rm -rf *.deb
	rm -rf *.rpm
	rm -rf dnsweb
	rm -rf ${DESTDIR} 

.PHONY: build

build:
	go build -o dnsweb  ${LDFLAGS}

amd64:
	GOOS=linux GOARCH=amd64 go build -o dnsweb  ${LDFLAGS}

.PHONY: pkg
pkg:
	rm -rf ${DESTDIR}
	mkdir ${DESTDIR}
	mkdir -p ${DESTDIR}/usr/local/bin
	mkdir -p ${DESTDIR}/usr/local/lib/dnsweb
	mkdir -p ${DESTDIR}/etc/
	cp dnsweb ${DESTDIR}/usr/local/bin/
	cp -r views ${DESTDIR}/usr/local/lib/dnsweb
	cp -r static ${DESTDIR}/usr/local/lib/dnsweb
	cp -r logrotate.d ${DESTDIR}/etc/
	mkdir -p ${DESTDIR}/etc/dnsweb/
	cp conf/app.conf  ${DESTDIR}/etc/dnsweb/app.conf
	mkdir -p ${DESTDIR}/etc/init.d/
	cp init.d/dnsweb ${DESTDIR}/etc/init.d/dnsweb

deb: clean amd64 pkg
	fpm -t deb -s dir -n dnsweb --rpm-os linux -v $(VERSION:v%=%) --config-files /etc/dnsweb --iteration ${RELEASE} -C ${DESTDIR}

rpm: clean amd64 pkg
	fpm -t rpm -s dir -n dnsweb --rpm-os linux -v ${VERSION} --config-files /etc/dnsweb --iteration ${RELEASE} -C ${DESTDIR}
