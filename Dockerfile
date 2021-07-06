FROM golang:alpine

RUN apk update && apk add xfsprogs xfsprogs-extra
RUN apk update && apk add util-linux

WORKDIR /xfs_project

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o xfs_quota
CMD ["./xfs_quota"]