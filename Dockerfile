FROM golang as builder
WORKDIR /code
ADD go.mod go.sum /code/
RUN go mod download
ADD . .
RUN go build -o /last_update_exporter main.go

FROM busybox:uclibc as busybox

FROM gcr.io/distroless/base-debian12
EXPOSE 9188
WORKDIR /
COPY --from=busybox /bin/sh /bin/sh
COPY --from=busybox /bin/wget /bin/wget
COPY --from=builder /last_update_exporter /usr/bin/last_update_exporter
ENTRYPOINT ["/usr/bin/last_update_exporter"]

