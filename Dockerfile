## BUILDER #####################################################################

FROM golang:alpine as builder

WORKDIR /go/src/github.com/essentialkaos/htmlcov

COPY . .

ENV GO111MODULE=auto

RUN apk add --no-cache git=~2.32 make=4.3-r0 upx=3.96-r1 && \
    make deps && \
    make all && \
    upx htmlcov

## FINAL IMAGE #################################################################

FROM essentialkaos/alpine:3.13

LABEL org.opencontainers.image.title="htmlcov" \
      org.opencontainers.image.description="Utility for converting coverage profiles into html pages" \
      org.opencontainers.image.vendor="ESSENTIAL KAOS" \
      org.opencontainers.image.authors="Anton Novojilov" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.url="https://kaos.sh/htmlcov" \
      org.opencontainers.image.source="https://github.com/essentialkaos/htmlcov"

COPY --from=builder /go/src/github.com/essentialkaos/htmlcov/htmlcov \
                    /usr/bin/

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["htmlcov"]

################################################################################
