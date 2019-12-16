# 构建后端可执行文件
# 这里基于golang镜像构建
FROM golang:1.13.0-alpine3.10 AS binary
# 拷贝源代码到镜像
ADD /go/src/crontab /go
# 为RUN、CMD、ENTRYPOINT、COPY和ADD设置工作目录
WORKDIR /go/src/crontab
# 构建镜像时运行的shell命令
RUN go mod tidy && \
    cd /go/master/main && \
    go build main.go
RUN cd /go/worker/main && \
    go build main.go
# 声明容器运行的服务端口
EXPOSE 8060
# 运行容器时执行的shell命令
CMD [".\master.go"]
CMD [".\waster.go"]
