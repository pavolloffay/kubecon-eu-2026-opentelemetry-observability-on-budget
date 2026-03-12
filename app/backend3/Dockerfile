FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -v -o /app

FROM scratch

COPY --from=builder /app /app

EXPOSE 5165

CMD ["/app"]
