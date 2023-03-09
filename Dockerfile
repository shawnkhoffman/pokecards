FROM golang:alpine AS build
WORKDIR /app
COPY . .
RUN go build -o pokecards .

FROM alpine
WORKDIR /app
COPY --from=build /app/pokecards .
VOLUME /app
ENTRYPOINT ["/app/pokecards"]
