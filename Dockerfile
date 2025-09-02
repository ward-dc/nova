FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Building with optimizatios for smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/main /main

EXPOSE 8080

ENTRYPOINT ["/main"]
