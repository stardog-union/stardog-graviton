FROM ubuntu:latest

RUN apt-get update
RUN apt-get install -y python python-pip zip curl ssh openjdk-9-jre

RUN curl -o terraform.zip https://releases.hashicorp.com/terraform/0.7.9/terraform_0.7.9_linux_amd64.zip
RUN unzip terraform.zip
RUN mv terraform /usr/local/bin

RUN curl -o packer.zip https://releases.hashicorp.com/packer/0.10.2/packer_0.10.2_linux_amd64.zip
RUN unzip packer.zip
RUN mv packer /usr/local/bin
