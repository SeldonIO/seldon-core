FROM rclone/rclone:1.57.0
LABEL name="Storage Initializer (rclone based)" \
      vendor="Seldon Technologies" \
      version="1.8.0-dev" \
      release="1" \
      summary="Storage Initializer for Seldon Core" \
      description="Allows Seldon Core to download artifacts from cloud and local storage to a local volume"


ENV RCLONE_CONFIG_GS_TYPE google cloud storage
ENV RCLONE_CONFIG_GS_ANONYMOUS true

# Do not set to '2' by default, as this may leak sensitive values
ENV RCLONE_VERBOSE 1

ENTRYPOINT ["rclone", "copy"]
