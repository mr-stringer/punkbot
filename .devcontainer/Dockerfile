FROM golang:1.24.0-bookworm

ARG USER='engineer'
ARG UID=5000
ARG GID=${UID}
ENV LANG=C.UTF-8

# Update and install needed packages
RUN apt-get -y update && \
  rm -rf /var/lib/apt/lists/*

# Add user
RUN groupadd --gid ${GID} ${USER} && \
    useradd -d /home/$USER --uid ${UID} --gid ${GID} -m -s /bin/bash $USER


CMD [ "sleep", "infinity" ]
