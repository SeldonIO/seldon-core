FROM rclone/rclone:1.54.0
LABEL name="Storage Initializer (rclone based)" \
      vendor="Seldon Technologies" \
      version="1.8.0-dev" \
      release="1" \
      summary="Storage Initializer for Seldon Core" \
      description="Allows Seldon Core to download artifacts from cloud and local storage to a local volume"

ENTRYPOINT ["rclone", "copy"]
