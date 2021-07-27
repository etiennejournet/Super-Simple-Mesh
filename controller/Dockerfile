FROM golang:1.16-alpine as builder
WORKDIR /go/src/ssm
COPY . .
RUN go install -v ./...

FROM alpine:3.13
LABEL TAG=0.4.0
ARG BIN_PATH=/usr/local/bin/
COPY --from=builder /go/bin/ $BIN_PATH

RUN apk --no-cache add ca-certificates
RUN adduser --disabled-password --gecos "" gouser  && chown -R gouser $BIN_PATH
USER gouser
CMD ["ssm"]
