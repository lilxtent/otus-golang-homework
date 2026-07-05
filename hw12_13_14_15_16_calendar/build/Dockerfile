FROM golang:1.26-alpine AS build

ARG SERVICE=calendar
ARG LDFLAGS

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o /opt/calendar/${SERVICE} ./cmd/${SERVICE}

FROM alpine:3.22

ARG SERVICE=calendar

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar"
LABEL MAINTAINERS="student@otus.ru"

RUN addgroup -S calendar && adduser -S calendar -G calendar

COPY --from=build /opt/calendar/${SERVICE} /opt/calendar/${SERVICE}
COPY ./configs /etc/calendar

USER calendar

ENV SERVICE=${SERVICE}
ENV CONFIG_FILE=/etc/calendar/config.yaml

ENTRYPOINT ["/bin/sh", "-c", "/opt/calendar/${SERVICE} --config=${CONFIG_FILE}"]
