FROM golang:1.26-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/service ./services/governance-svc/cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/service /service
EXPOSE 8105
USER nonroot:nonroot
ENTRYPOINT ["/service"]
