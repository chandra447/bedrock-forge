FROM golang:1.23-alpine

# Install git and ca-certificates
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy only necessary source code directories
COPY internal/ internal/
COPY cmd/ cmd/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 GOOS=linux go build -o bedrock-forge ./cmd/bedrock-forge

# Create entrypoint script directly in the image and put it in /usr/local/bin
RUN printf '#!/bin/sh\n\
set -e\n\
\n\
# Default values\n\
COMMAND=${INPUT_COMMAND:-"generate"}\n\
OUTPUT_DIR=${INPUT_OUTPUT_DIR:-"terraform"}\n\
CONFIG_PATH=${INPUT_CONFIG_PATH:-"."}\n\
\n\
# Build arguments array\n\
ARGS="$COMMAND"\n\
\n\
if [ "$COMMAND" = "generate" ]; then\n\
    ARGS="$ARGS $CONFIG_PATH $OUTPUT_DIR"\n\
elif [ "$COMMAND" = "validate" ] || [ "$COMMAND" = "scan" ]; then\n\
    ARGS="$ARGS $CONFIG_PATH"\n\
fi\n\
\n\
# Add validation config if provided\n\
if [ -n "$INPUT_VALIDATION_CONFIG" ]; then\n\
    ARGS="$ARGS --validation-config $INPUT_VALIDATION_CONFIG"\n\
fi\n\
\n\
# Add debug flag if enabled\n\
if [ "$INPUT_DEBUG" = "true" ]; then\n\
    ARGS="$ARGS --debug"\n\
fi\n\
\n\
echo "Executing: /app/bedrock-forge $ARGS"\n\
exec /app/bedrock-forge $ARGS\n' > /usr/local/bin/entrypoint.sh

RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
