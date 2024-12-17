FROM golang:1.22.7-alpine

ARG USERNAME=renatus
ARG USER_UID=1000

RUN adduser -u $USER_UID $USERNAME -D && \
    apk add make curl jq && \
    chown renatus:renatus -R /go && \
    go install github.com/pressly/goose/v3/cmd/goose@latest
ENV PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

USER $USERNAME
