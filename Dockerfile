FROM golang:1.26.5-trixie AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/awsping \
    ./cmd/awsping

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build --chown=nonroot:nonroot /out/awsping /awsping

USER nonroot:nonroot
ENTRYPOINT ["/awsping"]
