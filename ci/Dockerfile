FROM ubuntu

RUN apt-get update

RUN apt-get install -y python curl wget unzip make git groff openjdk-8-jdk zip awscli s3cmd python-boto
RUN wget https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz
RUN tar -C /usr/local/ -zxvf go1.8.1.linux-amd64.tar.gz
RUN wget https://releases.hashicorp.com/terraform/0.8.8/terraform_0.8.8_linux_amd64.zip
RUN unzip terraform_0.8.8_linux_amd64.zip
RUN mv terraform /usr/local/bin
RUN wget https://releases.hashicorp.com/packer/0.12.3/packer_0.12.3_linux_amd64.zip
RUN unzip packer_0.12.3_linux_amd64.zip
RUN mv packer /usr/local/bin
RUN mkdir -p /usr/local/src/go/src/github.com/stardog-union/

ENV PATH="/usr/local/src/go/bin:/usr/local/go/bin:${PATH}"
ENV GOPATH=/usr/local/src/go
RUN go get github.com/mitchellh/gox
