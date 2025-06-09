FROM golang:1.21 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/k8s-feature-reaper .

FROM gcr.io/distroless/static
COPY --from=builder /out/k8s-feature-reaper /k8s-feature-reaper
ENTRYPOINT ["/k8s-feature-reaper"]
