FROM tiangolo/meinheld-gunicorn-flask:python3.9
LABEL maintainer="Orchest B.V. https://www.orchest.io"

# Install required system packages and refresh certs
RUN apt-get update \
  && apt-get install -y ca-certificates git rsync netcat-traditional \
  && update-ca-certificates --fresh

# Install nodejs
ENV NODE_VERSION=18.18.2
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
ENV NVM_DIR=/root/.nvm
RUN . "$NVM_DIR/nvm.sh" && nvm install v${NODE_VERSION}
RUN . "$NVM_DIR/nvm.sh" && nvm use v${NODE_VERSION}
RUN . "$NVM_DIR/nvm.sh" && nvm alias default v${NODE_VERSION}
ENV PATH="$NVM_DIR/versions/node/v${NODE_VERSION}/bin/:${PATH}"

# Install the Python requirements before the JS compiling, because
# changes in the front-end often include JS changes meaning that its
# cache is more likely to be invalidated.
COPY ./requirements.txt /orchest/services/orchest-webserver/
COPY ./lib/python /orchest/lib/python

# Set the `WORKDIR` so the editable installs in the `requirements.txt`
# can use relative paths.
WORKDIR /orchest/services/orchest-webserver
RUN pip3 install --upgrade pip && pip3 install -r requirements.txt

# Static application files.
COPY ./client ./client
COPY ./lib/javascript /orchest/lib/javascript
COPY ./lib/design-system /orchest/lib/design-system

# PNPM files
COPY ./pnpm_files/* /orchest/


# Compile front-end code with npx and compile style.
WORKDIR /orchest

RUN npm run setup
RUN pnpm i --strict-peer-dependencies=false && \
  pnpm run build --stream --filter "@orchest/client-orchest" && \
  # Clean node_modules to reduce image size
  find . -name 'node_modules' -type d -prune -exec rm -rf '{}' +

# Application files.
WORKDIR /orchest/services/orchest-webserver/
COPY ./app ./app

# Setting this WORKDIR is required by the base image: "otherwhise gunicorn
# will try to run the app in /app". Additionally, we need to specify a
# custom path for the `gunicorn_conf.py` file.
WORKDIR /orchest/services/orchest-webserver/app

ENV GUNICORN_CONF /orchest/services/orchest-webserver/app/gunicorn_conf.py
ENV APP_MODULE "main:app"
ARG ORCHEST_VERSION
ENV ORCHEST_VERSION=${ORCHEST_VERSION}
COPY ./start.sh /
