FROM moby/buildkit:v0.10.4
LABEL maintainer="Orchest B.V. https://www.orchest.io"

# Overwrite the apk repositories to use Alpine v3.16
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.16/main" > /etc/apk/repositories \
    && echo "http://dl-cdn.alpinelinux.org/alpine/v3.16/community" >> /etc/apk/repositories \
    && apk update \
    && apk --no-cache upgrade openssh-client \
    && apk add rsync podman \
    && ssh-keygen -A \
    # To ssh from the jupyter server container while building to be
    # able to write to jupyter settings.
    && echo -e "root\nroot" | passwd root \
    && echo "PermitRootLogin yes" >> /etc/ssh/sshd_config

RUN buildkitd --version \
    && buildctl --version
