**模块三作业**

- 构建本地镜像

  

  - 编写 Dockerfile 将练习 2.2 编写的 httpserver 容器化

    - ```dockerfile
      # Dockerfile detail
      # pre_build stage
      FROM golang:1.17.1 as pre_build
      WORKDIR /
      COPY web.go /
      RUN go env -w GO111MODULE=auto && export GOPROXY=https://goproxy.cn,direct \
          && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o httpServer .
      
      # final stage
      FROM scratch
      COPY --from=pre_build /httpServer .
      ENTRYPOINT ["./httpServer"]
      ```

    - ```
      # build command
      cd $WORKDIR
      docker build ./
      ```

      

  - 将镜像推送至 docker 官方镜像仓库

    - ```
      # docker tag and push command
      docker tag $imageID hellodockerhello/httpserver:v1
      docker push hellodockerhello/httpserver:v1
      
      # docker image url
      https://hub.docker.com/repository/docker/hellodockerhello/httpserver
      ```

      

  - 通过 docker 命令本地启动 httpserver

    - ```
      # run container in detached mode (run in the background)
      docker run -d hellodockerhello/httpserver:v1
      ```

      

  - 通过 nsenter 进入容器查看 IP 配置

    - ```
      # get pid
      pid=`docker inspect a11ca233bc9fa122efc40cb41a766fdba5bc64a6ec4bf27ac6533ef1b9f1c301  -f '{{.State.Pid}}' `
      # use nsenter to enter network namespace
      nsenter -t $pid -n
      # use ifconfig to show network config
      ifconfig -a
      # exit from network namespace
      exit
      ```

      