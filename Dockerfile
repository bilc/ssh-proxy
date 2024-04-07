FROM ubuntu:20.04 
 
#for ssh proxy
#https://get.helm.sh/helm-v3.14.3-linux-amd64.tar.gz
COPY  linux-amd64/helm  /usr/local/bin
COPY proxy-k8s  /usr/local/bin
COPY /conf/proxy-k8s.yaml  /etc/

#for ssh server
RUN apt update 
RUN apt install -y openssh-server
RUN mkdir /run/sshd

#useradd -d /home/test -m test
#echo "user003:123456" | chpasswd
ENTRYPOINT [ "bash","-c", "/usr/sbin/sshd -D -e" ]
