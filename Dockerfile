FROM golang:1.7-wheezy

RUN go get -u \
      github.com/golang/dep/... \
      github.com/mitchellh/gox \
      github.com/tcnksm/ghr

WORKDIR /go/src/github.com/minodisk/nvim-finder-rplugin
COPY . .
RUN dep ensure && dep status
RUN sh build.sh
