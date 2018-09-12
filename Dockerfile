FROM debian:9.5-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY build/query-store-quay-logs /usr/local/bin/query-store-quay-logs

CMD ["/usr/local/bin/query-store-quay-logs"]