# pre_build stage
FROM golang:1.17.1 as pre_build
WORKDIR /
COPY web.go /
RUN go env -w GO111MODULE=auto && export GOPROXY=https://goproxy.cn,direct \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o httpServer .

# final stage
FROM scratch
COPY --from=pre_build /httpServer .
ENTRYPOINT ["./httpServer"]
