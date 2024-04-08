FROM ubuntu:22.04 
 
#for ssh proxy
#https://get.helm.sh/helm-v3.14.3-linux-amd64.tar.gz
COPY  linux-amd64/helm  /usr/local/bin

RUN mkdir -p /work
COPY proxy-k8s  /work
COPY /conf/proxy-k8s.yaml  /work
COPY simple-server-chart  /work/simple-server-chart

#for ssh server
RUN apt update 
RUN apt install -y openssh-server
RUN mkdir /run/sshd

#useradd -d /home/test -m test
#echo "user003:123456" | chpasswd
ENTRYPOINT [ "bash","-c", "/usr/sbin/sshd -D -e" ]
