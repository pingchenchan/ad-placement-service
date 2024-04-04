FROM golang:1.21 as builder


WORKDIR /app


COPY go.* ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -v -o app


FROM scratch
COPY --from=builder /app/app ./
COPY --from=builder /app/models/countryCode.json ./countryCode.json
CMD ["./app"]
