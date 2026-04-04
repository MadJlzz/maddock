FROM golang:1.26 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /maddock-agent ./cmd/agent/

FROM fedora:42

COPY --from=build /maddock-agent /usr/local/bin/maddock-agent
ENTRYPOINT ["maddock-agent"]
