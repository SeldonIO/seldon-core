FROM rclone/rclone:1.55.1
LABEL name="Storage Initializer (rclone based)" \
      vendor="Seldon Technologies" \
      version="1.8.0-dev" \
      release="1" \
      summary="Storage Initializer for Seldon Core" \
      description="Allows Seldon Core to download artifacts from cloud and local storage to a local volume"


ENV RCLONE_CONFIG_GS_TYPE google cloud storage
ENV RCLONE_CONFIG_GS_ANONYMOUS true

ENTRYPOINT ["rclone", "copy", "-v"]
