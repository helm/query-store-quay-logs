FROM alpine:3.8 as builder
RUN apk add --no-cache --virtual .build-deps ca-certificates

# Create appuser so as not to run as root
RUN adduser -D -g '' appuser

# The apps image can be from scratch rather than alpine
FROM scratch

# Copying the non-root user to the final image
COPY --from=builder /etc/passwd /etc/passwd

# The app needs the ca certificates to make calls our to Quay and Azure. Copying
# them over
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# The app was built outside the image process. Copy that in.
COPY build/query-store-quay-logs /query-store-quay-logs

USER appuser
CMD ["/query-store-quay-logs"]
