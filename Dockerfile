FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /birth-reminder ./cmd/birthday-reminder

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /birth-reminder /app/birth-reminder
USER 65532:65532
ENTRYPOINT ["/app/birth-reminder"]
