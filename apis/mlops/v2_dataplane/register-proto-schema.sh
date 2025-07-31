#!/bin/bash

# Schema Registry endpoint
SCHEMA_REGISTRY_URL="http://localhost:8081"

# Directory containing .proto files
PROTO_DIR="./schema_registry"

# Loop over all .proto files in the directory
for PROTO_FILE in "$PROTO_DIR"/*.proto; do
  if [[ -f "$PROTO_FILE" ]]; then
    # Extract base filename without extension for the subject name
    FILENAME=$(basename -- "$PROTO_FILE")
    SUBJECT_NAME="${FILENAME%.*}"

    echo "Registering $PROTO_FILE as subject $SUBJECT_NAME"

    # Read and escape schema
    SCHEMA=$(cat "$PROTO_FILE" | jq -Rs .)

    echo "the schema register would be $SCHEMA"

    # Post to Schema Registry
    curl -s -X POST "${SCHEMA_REGISTRY_URL}/subjects/${SUBJECT_NAME}/versions" \
      -H "Content-Type: application/vnd.schemaregistry.v1+json" \
      -d '{
            "schemaType": "PROTOBUF",
            "schema": '"${SCHEMA}"'
          }' \
      && echo "✓ Registered $SUBJECT_NAME" \
      || echo "✗ Failed to register $SUBJECT_NAME"
  fi
done