FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /mcp-memory-libravdb ./cmd/mcp-memory-libravdb

FROM scratch
COPY --from=build /mcp-memory-libravdb /mcp-memory-libravdb
EXPOSE 8082
ENTRYPOINT ["/mcp-memory-libravdb"]
CMD ["http"]
