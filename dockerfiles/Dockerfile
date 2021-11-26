FROM golang:1.17.3

RUN mkdir /workspace
WORKDIR /workspace/

ENV CGO_ENABLED 1

RUN apt update
RUN apt install -y pkg-config lxc-dev
RUN apt install -y python2.7 python2.7-dev
RUN ln -s /usr/bin/python2.7 /usr/bin/python

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go install github.com/go-delve/delve/cmd/dlv@latest
# RUN aptinstall pip
#RUN curl https://bootstrap.pypa.io/pip/2.7/get-pip.py --output get-pip.py
#RUN python get-pip.py

# RUN apt install -y build-essential

#docker run  -it -v C:\Users\lie\Desktop\cloudinsight-agent\:/workspace  agent-dev-image  bash
#docker run -itd -v  C:\Users\lie\Desktop\cloudinsight-agent\:/workspace  -p 2345:2345 --name debug  agent-dev-image  bash