#源镜像
FROM golang:1.10
#作者
MAINTAINER Razil "duan@isnb.net"
#设置工作目录
WORKDIR $GOPATH/src/github.com/isnbnet/k8sGo
#将服务器的go工程代码加入到docker容器中
ADD . $GOPATH/src/github.com/isnbnet/k8sGo
#go构建可执行文件
RUN go build .
#暴露端口
EXPOSE 80
#最终运行docker的命令
ENTRYPOINT  ["./k8sGo"]