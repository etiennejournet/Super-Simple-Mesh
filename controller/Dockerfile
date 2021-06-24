FROM golang:1.16-alpine as builder
WORKDIR /go/src/ssm
COPY . .
RUN go mod init ssm && go get -u ./... && go install -v ./...

FROM alpine:3.13
ARG BIN_PATH=/usr/local/bin/
COPY --from=builder /go/bin/ $BIN_PATH

RUN apk --no-cache add ca-certificates
RUN adduser --disabled-password --gecos "" gouser  && chown -R gouser $BIN_PATH
USER gouser
CMD ["ssm"]
