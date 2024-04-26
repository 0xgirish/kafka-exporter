FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Girish Kumar <git@0xgirish.in>

ARG TARGETARCH
ARG BIN_DIR=.build/linux-${TARGETARCH}/

COPY ${BIN_DIR}/kafka-exporter /bin/kafka-exporter

EXPOSE     9308
USER nobody
ENTRYPOINT [ "/bin/kafka-exporter" ]