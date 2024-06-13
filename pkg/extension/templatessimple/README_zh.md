## 前置条件

手动安装一个ingress控制器, 默认设置为NodePort模式,端口30888, 在您熟悉整个流程前, 建议不要调整

```bash
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx --create-namespace \
  --set controller.service.type=NodePort \
  --set controller.service.nodePorts.http=30888
```
## 创建一个扩展

使用一个已制作好的chart包,或者生成一个示例的

```bash
helm create demo
helm package demo
Successfully packaged chart and saved it to: /Users/inksnw/Desktop/demo-0.1.0.tgz
```

创建扩展

```bash
# --from 添加上文中的chart包
ksbuilder createsimple --from=./demo-0.1.0.tgz 
```

## 验证

> demo为chart包的名字,即chart.yaml中的name字段

### 域名方式

安装后, 您可以访问以下示例地址验证

- http://demo.www.ks.com:30888/  验证子域名解析是否正常

- http://www.ks.com:30880/pstatic/dist/demo/index.js 验证前端js代理是否正常

### nip.io

- http://demo.192.168.50.208.nip.io:30888/ 验证子域名解析是否正常

- http://192.168.50.208:30880/pstatic/dist/demo/index.js 验证前端js代理是否正常

