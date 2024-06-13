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

在应用商店可看到相应应用

