# Ubuntu 20.04.1 LTS Focal
FROM jupyter/base-notebook:2022-03-09
LABEL maintainer="Dadosfera USA Corporation https://app.dadosfera.ai/"

USER root

# defailt-libmysqlclient-dev : Orchest dependency
# libkrb5-dev                : Enterprise Gateway requirement
# RUN apt-get update \
#     && apt-get install -yq --no-install-recommends openssh-server git default-libmysqlclient-dev libkrb5-dev \
#     bzip2 \
#     libxcb1 \
#     libdbus-1-3 \
#     curl \
#     ca-certificates \
#     apt-get clean && \
#     && rm -rf /var/lib/apt/lists/*
RUN apt-get update \
    && apt-get install -yq --no-install-recommends curl openssh-server git default-libmysqlclient-dev libkrb5-dev libxcb1 \
    && rm -rf /var/lib/apt/lists/*

RUN passwd -d $NB_USER \
    # Enable `sudo` for the $NB_USER by default.
    && echo "$NB_USER   ALL=(ALL)   NOPASSWD:ALL" | tee /etc/sudoers.d/$NB_USER \
    # Persist the value of `DEBIAN_FRONTEND` when running with `sudo`,
    # so that installing packages works as expected.
    && echo "Defaults env_keep += \"DEBIAN_FRONTEND\"" | tee --append /etc/sudoers.d/$NB_USER \
    # All files in this directory should be mode 0440.
    && chmod 0440 /etc/sudoers.d/$NB_USER

# Get Enterprise Gateway kernel files.
WORKDIR /usr/local/bin
RUN wget https://github.com/jupyter-server/enterprise_gateway/releases/download/v2.5.2/jupyter_enterprise_gateway_kernel_image_files-2.5.2.tar.gz -O kernel_image_files.tar.gz \
    && tar -xf kernel_image_files.tar.gz \
    && rm -rf kernel_image_files.tar.gz kernel-launchers/scala \
    && chown jovyan:users /usr/local/bin/bootstrap-kernel.sh \
    && chmod 0755 /usr/local/bin/bootstrap-kernel.sh \
    && chown -R jovyan:users /usr/local/bin/kernel-launchers

# Install Enterprise Gateway requirements.
RUN mamba install --quiet --yes \
    cffi \
    ipykernel \
    ipython \
    'jupyter_client<7' \
    future \
    pycryptodomex && \
    mamba clean --all -f -y && \
    fix-permissions $CONDA_DIR && \
    fix-permissions /home/$NB_USER

# Get all Orchest requirements in place.
COPY ./runnable-shared/runner/requirements* /orchest/services/base-images/runnable-shared/runner/
COPY ./lib/python /orchest/lib/python
COPY ./orchest-sdk /orchest/orchest-sdk
RUN chown $NB_USER -R /orchest/orchest-sdk \
    && chown $NB_USER -R /orchest/lib && \
    echo -e "jovyan\njovyan" | passwd jovyan

USER $NB_USER

WORKDIR /orchest/services/base-images/runnable-shared/runner

# Install user requirements to be able to use Orchest. Install them in
# the `base` conda environment as it is default environment.
# NOTE: Use `pip` so we can install from local sources.
RUN pip install -r requirements-user.txt --no-cache

# Install Orchest dependencies in our own environment so they are
# completely isolated from user dependencies. Use a venv instead
# of conda environment because it is much smaller.
RUN python -m venv /home/$NB_USER/venv \
    && source /home/$NB_USER/venv/bin/activate \
    && pip install -r requirements.txt --no-cache \
    && deactivate

# Empty configuration file for kernel initialization
RUN touch /home/jovyan/.orchestrc && \
    echo "export PS1='\${ENVIRONMENT_SHELL_NAME:+(\$ENVIRONMENT_SHELL_NAME)}\[\033[01;32m\]\u\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ '" >> ~/.bashrc

# Copy application files as late as possible to avoid cache busting.
COPY ./runnable-shared/runner* /orchest/services/base-images/runnable-shared/runner

# Set the working directory
WORKDIR /orchest/services/base-images/base-kernel-py-agent
RUN CONFIGURE=false GOOSE_BIN_DIR=/usr/local/bin curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh | bash

# Add /root/.local/bin to the PATH

RUN echo "Installation and initial configuration of Goose completed!" \
    && echo "To start a session, run: goose session"


# This path is searched first to locate kernels. Without this variable
# Jupyter will search inside the orchestdependencies environment first
# and end up using the wrong executable to start the kernel.
ENV JUPYTER_PATH=/opt/conda/share/jupyter
ENV HOME=/home/$NB_USER
ENV BASH_ENV=/home/jovyan/.orchestrc


RUN pip3 install pyopenssl --upgrade
RUN --mount=type=secret,id=aws_access_key_id,uid=1000 \
    --mount=type=secret,id=aws_secret_access_key,uid=1000 \
    export AWS_ACCESS_KEY_ID=$(cat /run/secrets/aws_access_key_id) && \
    export AWS_SECRET_ACCESS_KEY=$(cat /run/secrets/aws_secret_access_key) && \
    export AWS_DEFAULT_REGION=us-east-1 && \
    pip3 install boto3 awscli chardet requests anybase32 'numpy<2' pandas snowflake-snowpark-python fastparquet streamlit && \
    aws codeartifact login --tool pip --domain dadosfera --domain-owner 611330257153 --region us-east-1 --repository dadosfera-pip && \
    pip3 install dadosfera==1.8.0b6 dadosfera_logs==1.0.3
RUN rm ~/.config/pip/pip.conf

RUN curl -fsSL https://code-server.dev/install.sh | sh -s --
# Orchest related environment variable that can be set to specify the
# conda environment to use to start Jupyter kernels.
ENV CONDA_ENV="base"

# Set a default renderer for plotly that actually renders in JupyterLab.
ENV PLOTLY_RENDERER="iframe"


# Required by the Enterprise Gateway to launch an ipykernel.
ENV KERNEL_LANGUAGE="python"

ARG ORCHEST_VERSION
ENV ORCHEST_VERSION=${ORCHEST_VERSION}

RUN echo PATH="${HOME}/.local/bin:${PATH}" >> ${HOME}/.orchestrc
RUN echo PATH="${HOME}/.local/bin:${PATH}" >> ${HOME}/.bashrc

COPY ./runnable-shared/bootscript.sh /orchest/bootscript.sh
CMD [ "/orchest/bootscript.sh" ]
