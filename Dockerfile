FROM golang:alpine

RUN apk add --no-cache xfsprogs

WORKDIR /xfs_project

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o xfs_quota
CMD ["./xfs_quota"]