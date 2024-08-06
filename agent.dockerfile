FROM golang:1.22 AS builder

ENV CGO_ENABLED=0

WORKDIR /builder

COPY . .
RUN go build -o agent cmd/agent/main.go

FROM alpine

ENV PORT=8081

WORKDIR /app

COPY --from=builder /builder/agent .

ENTRYPOINT ["./agent"]