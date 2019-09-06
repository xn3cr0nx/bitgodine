FROM golang:1.13 as builder
LABEL maintainer="Patrick Jusic <patrick.jusic@protonmail.com>"

WORKDIR /bitgodine
COPY . .

ARG github_token
ENV github_token=$github_token

RUN git config \
  --global \
  url."https://${github_token}:x-oauth-basic@github.com/xn3cr0nx".insteadOf \
  "https://github.com/xn3cr0nx"

RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/bitgodine -v ./cmd/bitgodine

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /bitgodine/build/bitgodine .
EXPOSE 3000
CMD ["./bitgodine", "serve", "--dgHost", "dgraph_server"]
