FROM golang:1.16.5 as builder
LABEL maintainer="Patrick Jusic <patrick.jusic@protonmail.com>"

WORKDIR /spider

ENV GOOS=linux \
  GO111MODULE=on \
  GOPRIVATE="github.com/xn3cr0nx"
ARG GITHUB_TOKEN

RUN git config \
  --global \
  url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/xn3cr0nx".insteadOf \
  "https://github.com/xn3cr0nx"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/spider -v ./cmd/spider

FROM golang:alpine

WORKDIR /root/
RUN mkdir /root/.bitgodine
COPY --from=builder /spider/config ./config
COPY --from=builder /spider/build/spider .
# CMD ["./spider", "crawl"]

