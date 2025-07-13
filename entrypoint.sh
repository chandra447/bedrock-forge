#!/bin/sh
set -e

# Default values
COMMAND=${INPUT_COMMAND:-"generate"}
OUTPUT_DIR=${INPUT_OUTPUT_DIR:-"terraform"}
CONFIG_PATH=${INPUT_CONFIG_PATH:-"."}

# Build arguments array
ARGS="$COMMAND"

if [ "$COMMAND" = "generate" ]; then
    ARGS="$ARGS $CONFIG_PATH $OUTPUT_DIR"
elif [ "$COMMAND" = "validate" ] || [ "$COMMAND" = "scan" ]; then
    ARGS="$ARGS $CONFIG_PATH"
fi

# Add validation config if provided
if [ -n "$INPUT_VALIDATION_CONFIG" ]; then
    ARGS="$ARGS --validation-config $INPUT_VALIDATION_CONFIG"
fi

# Add debug flag if enabled
if [ "$INPUT_DEBUG" = "true" ]; then
    ARGS="$ARGS --debug"
fi

echo "Executing: ./bedrock-forge $ARGS"
exec ./bedrock-forge $ARGS