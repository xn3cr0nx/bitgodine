FROM golang:1.12 as builder
LABEL maintainer="Patrick Jusic <patrick.jusic@protonmail.com>"

WORKDIR /bitgodine
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./build/bitgodine -v ./cmd/bitgodine

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /bitgodine/build/bitgodine .
CMD ["./bitgodine", "serve", "--dgHost", "dgraph_server"]