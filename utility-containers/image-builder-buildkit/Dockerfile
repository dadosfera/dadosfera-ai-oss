FROM moby/buildkit:v0.10.4
LABEL maintainer="Orchest B.V. https://www.orchest.io"
RUN buildkitd --version \
    && buildctl --version \
    && apk update \
    && apk --no-cache upgrade openssh-client \
    && apk add rsync \
    && ssh-keygen -A \
    # To ssh from the jupyter server container while building to be
    # able to write to jupyter settings.
    && echo -e "root\nroot" | passwd root \
    && echo "PermitRootLogin yes" >> /etc/ssh/sshd_config
