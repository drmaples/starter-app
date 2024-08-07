FROM drmaples/starter-app/app-builder AS builder

FROM scratch

LABEL org.opencontainers.image.source=https://github.com/drmaples/starter-app/migrate

ARG APP_VERSION
ARG BUILD_DATE
ARG COMMIT_HASH

ENV BUILD_DATE=${BUILD_DATE}
ENV COMMIT_HASH=${COMMIT_HASH}
ENV DD_VERSION=${APP_VERSION}
ENV SENTRY_RELEASE=${APP_VERSION}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/migrate .

# dont run as root in a container:
# https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#user
COPY --from=builder /etc/passwd /etc/passwd
USER scratchuser

EXPOSE 8000
CMD ["./migrate"]
