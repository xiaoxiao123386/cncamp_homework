- 模块 8 作业(第二部分)

  - 第一部分（见 **httpserver.yaml**）【直接跳到第二部分观看，第一部分忽略】

    - 优雅启动

      - 使用 readinessProbe 保证 Pod 启动时服务已经是就绪状态

        ```yaml
        # regularly probe http healthz uri to check whether service is normal
        readinessProbe:
        	httpGet:
        		path: /healthz
        		port: 80
        	initialDelaySeconds: 3
        	periodSeconds: 5
        	
        # 顺便写了个 postStart，但其实没太大用处，因为 liveness 及 readiness 都没用到这个写入的文件做为判断依据
        # when pod is started, echo info to a file  
        postStart:
        	exec:
        		command: ["/bin/sh", "-c", "echo 'start success' > /tmp/success_info"]
        ```

        

    - 优雅终止

      - 代码上，增加优雅终止的处理（使用 os/signal ）

        ```go
        // 优雅退出
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
        s := <-c
        log.Infof("Receive Signal [%s],Exit Properly\n", s) # 引入 logrus 打印日志
        ```

      - yaml 文件中，增加 preStop 及 terminationGracePeriodSeconds

        ```yaml
        # when pod is terminated, execute `killall -2` to ensure process exit properly
        preStop:
        	exec:
          	command: ["/bin/sh", "-c", "while killall -2 web; do sleep 1; done"]
          	
        # time to wait before changing a TERM signal to a KILL signal to the pod's main process
          terminationGracePeriodSeconds: 10
        ```

        

    - 资源需求和 QoS 保证

      - 配置 Pod 的 CPU 及内存需求，同时 QoS 策略也随之设置（变为 Burstable）

        ```yaml
        # set pod's resources requests and limits, it'll need at least 64MiB Memory and 0.5 CPU
        # and can't exceed 128MiB Memory and 1 CPU.
        # when set like this, its QoS type is "Burstable.
        resources:
        	requests:
        		memory: "64Mi"
        		cpu: "500m"
        	limits:
        		memory: "128Mi"
        		cpu: "1000m"
        ```

        

    - 探活

      - 使用 livenessProbe 来探测服务监听端口是否正常，以此来判断服务是否正常完成监听

        ```yaml
        # regularly probe tcp port 80 to check whether port 80 is listened correctly
        livenessProbe:
        	tcpSocket:
        		port: 80
          initialDelaySeconds: 3
          periodSeconds: 5
        ```

        

    - 日常运维需求，日志等级

      - 代码引入 logrus 打印日志，体现不同等级的日志输出

        ```go
        import (
        	...
        	log "github.com/sirupsen/logrus"
        	...
        )
        
        ...
        log.Info("http server start.")
        ...
        log.Fatal(err)
        ...
        log.Infof("Receive Signal [%s],Exit Properly\n", s)
        ```

        

    - 配置和代码分离（将 configmap 中对应键值作为环境变量注入，最终给代码返回 VERSION 作输入用）

      - 配置 configmap（见 http-config.yaml）

        ```yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: http-config
        data:
          env-parameters: "I'm httpserver's Env Parameters"
        ```

      - Pod 的 yaml 中，将 configmap 中的键为 env-parameters 的值作为 VERSION 的值注入

        ```yaml
        # inject env VERSION from configmap http-config
        env:
          - name: VERSION
        		valueFrom:
        			configMapKeyRef:
        				name: http-config
        				key: env-parameters
        ```

- 第二部分

  - 在第一部分的基础上提供更加完备的部署 spec，包括（不限于）：

    - Service

      - 思路：Pod 的 label 是 app: httpserver，因此 service 的 label selector 就保持一致，能选到对应后端 Pod 就行

        ```yaml
        apiVersion: v1
        kind: Service
        metadata:
          name: httpserver
        spec:
          selector:
            app: httpserver
          ports:
            - protocol: TCP
              port: 80
              targetPort: 80
        ```

        

    - Ingress

      - 第一步：先部署一个 nginx-controller，参照 `module8/module8_part2/nginx-ingress-deployment.yaml` —— 开源的也可以，这里直接用 cncamp101 中的 nginx-ingress-deployment.yaml 来做

        ```shell
        # 在集群中部署 nginx-ingress-controller
        kubectl apply -f nginx-ingress-deployment.yaml
        ```

      - 第二步：编写 ingress.yaml

        ```yaml
        apiVersion: networking.k8s.io/v1
        kind: Ingress
        metadata:
          name: test-ingress
          annotations:
            # use annotation, specify ingress controller to nginx ingress controller
            kubernetes.io/ingress.class: "nginx"
        spec:
          # config tls secret and host, the secret is what we create previously
          tls:
            - hosts:
              - cncamp.com
              secretName: cncamp-tls
          rules:
          # config host rules, when receive specified host and prefix uri, will transfer to backend service httpserver:80
          - host: cncamp.com
            http:
              paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: httpserver
                    port:
                      number: 80
        ```

      - 第三步，apply，获取 ingress ip、port，然后实际验证访问情况

        ```shell
        # apply ingress.yaml
        kubectl apply -f ingress.yaml
        
        # get ingress ip, from ADDRESS section, record $ADDRESS
        kubectl get ingress 
        
        # get node port, from PORT(S) section 443:NodePort/TCP, record $NodePort 
        kubectl get svc -n ingress-nginx
        
        # set hosts, cncamp.com to $ADDRESS
        echo $ADDRESS cncamp.com >> /etc/hosts
        
        # use curl tool to visit https://cncamp.com:$NodePort/healthz (returnEnv ...)
        curl https://cncamp.com:$NodePort/healthz -k
        ```

        

    可以考虑的细节

    - 如何确保整个应用的高可用。

      - 思路：采用 deployment 来保证高可用，这里采用三副本来做

        详见`module8/module8_part2/httpserver-deployment.yaml`
        
        `kubectl apply -f httpserver-deployment.yaml`

    - 如何通过证书保证 httpServer 的通讯安全。

      - 思路：参照 module8 中 ingress 文件夹的方式，先后创建密钥、证书及 secret
    
        ```shell
        # 命令如下
        # 生成证书及密钥
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=cncamp.com/O=cncamp" -addext "subjectAltName = DNS:cncamp.com"
        # 根据文件生成 tls secret
        kubectl create secret tls cncamp-tls --cert=./tls.crt --key=./tls.key
        ```
    
        
