FROM golang:1.27-alpine AS build

WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod ./
RUN go mod download

COPY main.go ./
COPY web/ ./web/

RUN go build -trimpath -ldflags="-s -w" -o /out/ipv6test .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/ipv6test /app/ipv6test

ENV LISTEN=:8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/ipv6test"]