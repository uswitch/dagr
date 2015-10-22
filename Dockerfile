FROM debian:jessie

RUN apt-get update && apt-get install -y --no-install-recommends ssh git python3 python3-pip

# Install AWS CLI
RUN pip3 install --upgrade pip
RUN pip3 install awscli

# Setup private key for accessing private repos
ADD id_rsa /root/.ssh/id_rsa
RUN chmod 600 /root/.ssh/id_rsa
RUN ssh-keyscan -t rsa github.com >> /root/.ssh/known_hosts

COPY dagr /usr/local/bin/dagr
RUN chmod +x /usr/local/bin/dagr

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY ui.tgz /go/src/app/ui.tgz
RUN tar -xzvf ui.tgz

CMD ["dagr"]

