FROM golang:1.21-alpine as build-base

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o ./out/assessment-tax .

FROM alpine:3.16.2

COPY --from=build-base /app/out/assessment-tax /app/assessment-tax

COPY .env .

CMD ["app/assessment-tax"]
