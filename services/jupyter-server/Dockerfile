FROM python:3.9-slim
LABEL maintainer="Orchest B.V. https://www.orchest.io"

RUN apt-get update \
    # Refresh SSL certificates
    && update-ca-certificates --fresh \
    && apt-get install -yq --no-install-recommends curl ssh sshpass git jq \
    rsync ca-certificates \
    # Install nodejs for Jupyter extensions.
    && curl -sL https://deb.nodesource.com/setup_lts.x | bash - \
    && apt-get install -yq --no-install-recommends nodejs \
    # Swap dash for bash to get more full fledged shell by default in
    # JupyterLab terminal.
    && ln -fs /bin/bash /bin/sh \
    && rm -rf /var/lib/apt/lists/*

# Build JupyterLab, pre-install thrid-party extensions and our custom
# extensions.
RUN mkdir -p /orchest/service/jupyter-server
WORKDIR /orchest/service/jupyter-server

COPY ./requirements.txt .
COPY ./lib/python /orchest/lib/python

ARG extension_dir=/jupyterlab-orchest-build
ARG pip_jupyter_extension_dir=/usr/local/share/jupyter/lab/extensions
RUN pip install pip --upgrade \
    # JupyterLab and possible extensions need to be installed first so
    # that they are included in the build.
    && pip install --no-cache -r requirements.txt \
    && mkdir -p $extension_dir \
    && cp -rfT $pip_jupyter_extension_dir $extension_dir/extensions \
    && jupyter lab build --dev-build=False --LabBuildApp.app_dir=$extension_dir \
    && jupyter lab clean --LabCleanApp.app_dir=$extension_dir

# Display warning message when running `pip` or `pip3` in the Jupyter
# terminal.
COPY patch-pip.sh pip-warning.py /
RUN bash /patch-pip.sh

COPY ./jupyter_server_config.py /root/.jupyter/jupyter_server_config.py
COPY shellspawner/shellspawner /usr/bin/shellspawner

# Add shellspawner to terminal profile
RUN echo "shellspawner --help && shellspawner" >> ~/.profile

# Hardcoded /project-dir because Python config can't be injected into Dockerfile
WORKDIR /project-dir

ARG ORCHEST_VERSION
ENV ORCHEST_VERSION=${ORCHEST_VERSION}

COPY start.sh /
ENTRYPOINT [ "/start.sh" ]
