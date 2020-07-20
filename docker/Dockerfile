# Build Geth in a stock Go builder container
FROM golang:1.14.3-stretch as builder

COPY sources.list /etc/apt/sources.list

RUN  apt-get update && apt-get install -y --force-yes git perl-base  curl bash cmake openssl make gcc g++ 

# ADD . /PlatONE-Go
#RUN cd /go-ethereum && make geth

# Pull Geth into a second stage deploy alpine container
#FROM alpine:latest

#RUN apk add --no-cache ca-certificates
#COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

#EXPOSE 8545 8546 30303 30303/udp
#ENTRYPOINT ["geth"]
