FROM golang:1.15 as build-env
RUN mkdir /app
WORKDIR /app
COPY . .
RUN go run main.go build

# It uses CGO (sqlite) so it is easier this way :D
FROM golang:1.15
COPY --from=build-env /app/elobot .
USER 1001
ENTRYPOINT ["./elobot"]
