ARG PG_VERSION

FROM golang:1.22-alpine as builder
ADD ./ /src
WORKDIR /src
RUN mkdir /dist && go build -o /dist/db-backup ./cmd/db-backup/

FROM postgres:${PG_VERSION}-alpine
RUN mkdir /backup
COPY --from=builder /dist /backup
WORKDIR /backup
RUN chmod 777 db-backup
ENTRYPOINT ["./db-backup"]