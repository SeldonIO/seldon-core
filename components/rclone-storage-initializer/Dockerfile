FROM rclone/rclone:1.59.1 as builder

RUN mkdir /licenses && wget -O /licenses/license.txt https://raw.githubusercontent.com/rclone/rclone/master/COPYING

FROM registry.access.redhat.com/ubi8/ubi-minimal
LABEL name="Rclone Storage Initializer" \
      vendor="Seldon Technologies" \
      version="1.15.0-dev" \
      release="1" \
      summary="Storage Initializer for Seldon Core" \
      description="Allows Seldon Core to download artifacts from cloud and local storage to a local volume"

COPY --from=builder /usr/local/bin/ /usr/local/bin/
COPY --from=builder /licenses/ /licenses/

ENV RCLONE_CONFIG_GS_TYPE google cloud storage
ENV RCLONE_CONFIG_GS_ANONYMOUS true

# Do not set to '2' by default, as this may leak sensitive values
ENV RCLONE_VERBOSE 1

USER 8888
ENTRYPOINT ["rclone", "copy"]
