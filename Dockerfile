FROM golang:alpine AS build
# add gcc tools
RUN apk add build-base
# run build process in /app directory
WORKDIR /app
# copy dependencies and get them
COPY ./go.mod ./go.sum ./
RUN go get -d -v ./...
# copy go src file(s)
COPY ./cmd ./cmd
COPY ./internal ./internal
# build the binary
RUN GOOS=linux CGO_ENABLED=1 GOARCH=amd64 go build -o notifi ./cmd/api/main.go

FROM alpine
# copy the binary from build stage
COPY --from=build /app/notifi /bin/notifi
# use non root
USER 1000:1000
# start server
CMD ["./bin/notifi"]
