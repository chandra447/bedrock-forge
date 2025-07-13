FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bedrock-forge ./cmd/bedrock-forge

FROM alpine:latest
RUN apk --no-cache add ca-certificates git
WORKDIR /root/

COPY --from=builder /app/bedrock-forge .

# Create entrypoint script inline to avoid file copying issues
RUN echo '#!/bin/sh' > entrypoint.sh && \
    echo 'set -e' >> entrypoint.sh && \
    echo '' >> entrypoint.sh && \
    echo '# Default values' >> entrypoint.sh && \
    echo 'COMMAND=${INPUT_COMMAND:-"generate"}' >> entrypoint.sh && \
    echo 'OUTPUT_DIR=${INPUT_OUTPUT_DIR:-"terraform"}' >> entrypoint.sh && \
    echo 'CONFIG_PATH=${INPUT_CONFIG_PATH:-"."}' >> entrypoint.sh && \
    echo '' >> entrypoint.sh && \
    echo '# Build arguments array' >> entrypoint.sh && \
    echo 'ARGS="$COMMAND"' >> entrypoint.sh && \
    echo '' >> entrypoint.sh && \
    echo 'if [ "$COMMAND" = "generate" ]; then' >> entrypoint.sh && \
    echo '    ARGS="$ARGS $CONFIG_PATH $OUTPUT_DIR"' >> entrypoint.sh && \
    echo 'elif [ "$COMMAND" = "validate" ] || [ "$COMMAND" = "scan" ]; then' >> entrypoint.sh && \
    echo '    ARGS="$ARGS $CONFIG_PATH"' >> entrypoint.sh && \
    echo 'fi' >> entrypoint.sh && \
    echo '' >> entrypoint.sh && \
    echo '# Add validation config if provided' >> entrypoint.sh && \
    echo 'if [ -n "$INPUT_VALIDATION_CONFIG" ]; then' >> entrypoint.sh && \
    echo '    ARGS="$ARGS --validation-config $INPUT_VALIDATION_CONFIG"' >> entrypoint.sh && \
    echo 'fi' >> entrypoint.sh && \
    echo '' >> entrypoint.sh && \
    echo '# Add debug flag if enabled' >> entrypoint.sh && \
    echo 'if [ "$INPUT_DEBUG" = "true" ]; then' >> entrypoint.sh && \
    echo '    ARGS="$ARGS --debug"' >> entrypoint.sh && \
    echo 'fi' >> entrypoint.sh && \
    echo '' >> entrypoint.sh && \
    echo 'echo "Executing: ./bedrock-forge $ARGS"' >> entrypoint.sh && \
    echo 'exec ./bedrock-forge $ARGS' >> entrypoint.sh && \
    chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]