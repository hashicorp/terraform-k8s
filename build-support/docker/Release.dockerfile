# This Dockerfile creates a production release image for the project. This
# downloads the release from releases.hashicorp.com and therefore requires that
# the release is published before building the Docker image.
#
# We don't rebuild the software because we want the exact checksums and
# binary signatures to match the software and our builds aren't fully
# reproducible currently.
FROM hashicorp/terraform:0.12.24

# NAME and VERSION are the name of the software in releases.hashicorp.com
# and the version to download. Example: NAME=terraform VERSION=1.2.3.
ARG NAME
ARG VERSION

# Set ARGs as ENV so that they can be used in ENTRYPOINT/CMD
ENV NAME=$NAME
ENV VERSION=$VERSION

# This is the location of the releases.
ENV HASHICORP_RELEASES=https://github.com/parkside-securities/$NAME/releases/download

# Create a non-root user to run the software.
RUN addgroup ${NAME} && \
    adduser -S -G ${NAME} ${NAME}

# Set up certificates, base tools, and software.
RUN set -eux && \
    apk add --no-cache ca-certificates curl gnupg libcap openssl su-exec iputils && \
    wget ${HASHICORP_RELEASES}/${VERSION}/${NAME}_${VERSION}_linux_amd64.zip && \
    wget ${HASHICORP_RELEASES}/${VERSION}/${NAME}_${VERSION}_SHA256SUMS && \
    unzip -d /bin ${NAME}_${VERSION}_linux_amd64.zip && \
    cd /tmp && \
    rm -rf /tmp/build && \
    apk del gnupg openssl && \
    rm -rf /root/.gnupg

USER ${NAME}
ENTRYPOINT ["${NAME}"]
