.PHONY: all prepare clean

DEB_TARGET_ARCH ?= armel

ifeq ($(DEB_TARGET_ARCH),armel)
GO_ENV := GOARCH=arm GOARM=5 CC_FOR_TARGET=arm-linux-gnueabi-gcc CC=$$CC_FOR_TARGET CGO_ENABLED=1
endif
ifeq ($(DEB_TARGET_ARCH),amd64)
GO_ENV := GOARCH=amd64 CC=x86_64-linux-gnu-gcc
endif
ifeq ($(DEB_TARGET_ARCH),i386)
GO_ENV := GOARCH=386 CC=i586-linux-gnu-gcc
endif

all: clean wb-shtrixmqtt

clean:
	rm -rf wb-shtrixmqtt

amd64:
	$(MAKE) DEB_TARGET_ARCH=amd64

wb-shtrixmqtt: main.go *.go
	$(GO_ENV) glide install
	$(GO_ENV) go build

install:
	mkdir -p $(DESTDIR)/usr/bin/ $(DESTDIR)/etc/init.d/ $(DESTDIR)/etc/wb-shtrixmqtt/ $(DESTDIR)/usr/share/wb-mqtt-confed/schemas $(DESTDIR)/etc/wb-configs.d $(DESTDIR)/usr/share/wb-shtrixmqtt-system/scripts/ $(DESTDIR)/usr/share/wb-shtrixmqtt/
	install -m 0755 wb-shtrixmqtt $(DESTDIR)/usr/bin/
	install -m 0660 wb-shtrixmqtt-conf.yaml $(DESTDIR)/etc/
	install -m 0755 initscripts/wb-shtrixmqtt $(DESTDIR)/etc/init.d/wb-shtrixmqtt


deb:
	CC=arm-linux-gnueabi-gcc dpkg-buildpackage -b -aarmel -us -uc
