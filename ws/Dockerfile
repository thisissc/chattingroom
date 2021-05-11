FROM alpine:latest
MAINTAINER Samuel <thisissc@gmail.com>

# ca-certificates fix "x509: failed to load system roots and no roots provided"
# tzdata fix "time: missing Location in call to Time.In"
RUN apk update \
  && apk add ca-certificates tzdata \
  && rm -rf /var/cache/apk/* \
  && adduser -D -g '' ecs-user

WORKDIR /opt/app/go/bin
USER ecs-user

COPY chat-ws chat-ws
COPY config.toml config.toml
ENTRYPOINT ["./chat-ws"]
