FROM node:alpine as node-builder
WORKDIR /workspace
ADD . /workspace
RUN npm install && npm run build:prod

FROM golang:alpine as go-builder
WORKDIR /workspace
ADD . /workspace
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix \
    cgo -o mail ./cmd/main.go

FROM scratch
WORKDIR /workspace
COPY --from=node-builder /workspace/public /workspace/public
COPY --from=go-builder   /workspace/mail   /workspace/mail

CMD ["/workspace/mail"]
