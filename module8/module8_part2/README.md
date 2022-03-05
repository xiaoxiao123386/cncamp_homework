- 最终的部署步骤

  ```shell
  # 部署 deployment
  kubectl apply -f httpserver-deployment.yaml
  kubectl apply -f http-config.yaml # 将 Pod 需要的 configmap 加上
  
  # 部署 service，关联刚才部署的 Pod
  kubectl apply -f httpserver-service.yaml
  
  # 部署 nginx ingress controller，为 ingress 生效做准备
  kubectl apply -f nginx-ingress-deployment.yaml
  
  # 生成 ingress 需要的 tls secret，最终 ingress 会用到这个 tls secret
  ## 生成证书及密钥，证书中域名设为 cncamp.com
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=cncamp.com/O=cncamp" -addext "subjectAltName = DNS:cncamp.com"
  ## 根据文件生成 tls secret
  kubectl create secret tls cncamp-tls --cert=./tls.crt --key=./tls.key
  
  # 部署 ingress，引用刚才生成的 tls secret，并将访问 cncamp.com、prefix 是 / 的请求转到前面配置的 service
  kubectl apply -f ingress.yaml
  
  
  # 获取访问 ingress 对外暴露的 ip、port，配置并最终访问
  ## get ingress ip, from ADDRESS section, record $ADDRESS
  kubectl get ingress 
  
  ## get node port, from PORT(S) section 443:NodePort/TCP, record $NodePort 
  kubectl get svc -n ingress-nginx
  
  ## set hosts, cncamp.com to $ADDRESS
  echo $ADDRESS cncamp.com >> /etc/hosts
  
  ## use curl tool to visit https://cncamp.com:$NodePort/healthz (returnEnv ...)
  curl https://cncamp.com:$NodePort/healthz -k # 从 ingress 进来，最终访问到 Pod 的 healthz 路径
  curl https://cncamp.com:$NodePort/returnHeader -k # 从 ingress 进来，最终访问到 Pod 的 returnHeader 路径
  ```

  
