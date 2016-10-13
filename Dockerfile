FROM golang
MAINTAINER Aleksei Melnik <alexshnup@gmail.com>
RUN mkdir -p /opt/gopath/src/github.com/alexshnup/wb-shtrixmqtt
ADD . /opt/gopath/src/github.com/alexshnup/wb-shtrixmqtt
WORKDIR /opt/gopath/src/github.com/alexshnup/wb-shtrixmqtt
ENV GOPATH=/opt/gopath
RUN go get github.com/alexshnup/go-config-manager/yaml && go get github.com/alexshnup/mqtt && go get github.com/alexshnup/yaml.v2 && go get github.com/alexshnup/uuid && go get golang.org/x/net/websocket && go get golang.org/x/text/encoding/charmap && go get github.com/sigurn/crc8 && ls
RUN go build
EXPOSE 9999/UDP
CMD [ "/opt/gopath/src/github.com/alexshnup/wb-shtrixmqtt/shtrixmqtt" ]
