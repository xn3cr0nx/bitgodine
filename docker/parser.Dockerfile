FROM golang:1.14.4 as builder 
LABEL maintainer="Patrick Jusic <patrick.jusic@protonmail.com>"

WORKDIR /bitgodine

COPY . .

RUN git config \
  --global \
  url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/xn3cr0nx".insteadOf \
  "https://github.com/xn3cr0nx"

RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/bitgodine -v ./cmd/bitgodine

FROM golang:alpine

WORKDIR /root/
RUN mkdir /root/.bitgodine
COPY --from=builder /bitgodine/build/bitgodine .
# CMD ["./bitgodine", "start", "--debug", "-r=false","--dgHost", "dgraph_server", "--blocksDir", "/bitcoin", "--dbDir", "/badger", "--utxo", "/bolt/utxoset.db"]
