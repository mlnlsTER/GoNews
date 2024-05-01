FROM golang:latest AS compiling_stage
RUN mkdir -p /go/src/censorship
WORKDIR /go/src/censorship
ADD server.go .
ADD go.mod .
RUN go install .
 
FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Zhdan Baliuk<balyuk603@gmail.com>"
WORKDIR /root/
COPY --from=compiling_stage /go/bin/censorship .
ENTRYPOINT ./censorship