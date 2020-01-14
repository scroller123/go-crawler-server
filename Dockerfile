FROM ubuntu:12.04

RUN apt-get update  
RUN apt-get install -y python-software-properties

RUN apt-get update  
RUN apt-get -y upgrade
RUN apt-get install -y libxml2 libxml2-dev pkg-config lxc-dev git
#golang-go
#RUN apt-get update

#RUN  add-apt-repository ppa:gophers/go
#RUN add-apt-repository ppa:hnakamur/golang-1.10
#RUN add-apt-repository ppa:duh/golang
#RUN add-apt-repository ppa:gophers/archive
#RUN apt-get update
#RUN apt-get install -y golang
#RUN apt-get install -y golang-stable 

RUN wget https://dl.google.com/go/go1.10.1.linux-amd64.tar.gz
RUN tar -xvf go1.10.1.linux-amd64.tar.gz
RUN mv go /usr/local
RUN export GOROOT=/usr/local/go
RUN export GOPATH=/var/server
RUN export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# ADD server.go /var/server/server.go
# ADD . /var/server
WORKDIR /var/server

#EXPOSE 3333

RUN /usr/local/go/bin/go version
RUN /usr/local/go/bin/go get github.com/PuerkitoBio/goquery

#CMD ["/usr/local/go/bin/go", "run", "/var/server/server.go"]  