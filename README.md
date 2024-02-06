# nodelocaldns-admission-webhook

自动将 DNSConfig 注入新建的 Pod，配合 NodeLocalDNS 使用，用户无需手动进行 DNS 配置。
给 pod 资源注入如下的 dnsPolicy 和 dnsConfig：
```yaml
dnsPolicy: none
dnsConfig:
nameservers:
- 169.254.20.10
- 10.96.0.10
searches:
- <namespace>.svc.cluster.local
- svc.cluster.local
- cluster.local
options:
- name: ndots
  value: "3"
- name: attempts
  value: "2"
- name: timeout
  value: "1"
```

通过 admission webhook 方案，避免了修改节点上 kubelet 参数或者手动修改 Yaml。

## 规则

Pod在同时满足以下条件时，才会自动注入DNS缓存。如果您的Pod容器未注入DNS缓存服务器的IP地址，请检查Pod是否未满足以下条件。

1. 新建Pod不位于kube-system和kube-public命名空间。

2. 新建Pod所在命名空间的Labels标签包含node-local-dns-injection=enabled。

3. 新建Pod没有被打上禁用DNS注入node-local-dns-injection=disabled标签。

4. 新建Pod的网络为hostNetwork且DNSPolicy为ClusterFirstWithHostNet，或Pod为非hostNetwork且DNSPolicy为ClusterFirst。

Pod will automatically inject DNS cache when all of the following conditions are met:
1. The newly created Pod is not in the kube-system and kube-public namespaces.
2. The Labels of the namespace where the new Pod is located contain node-local-dns-injection=enabled.
3. The newly created Pod is not labeled with the disabled DNS injection node-local-dns-injection=disabled label.
4. The network of the newly created Pod is hostNetwork and DNSPolicy is ClusterFirstWithHostNet, or the Pod is non-hostNetwork and DNSPolicy is ClusterFirst.

## 部署

在 deploy 目录提供了部署相关 yaml，apply 即可。

* 1）部署 cert-manager 用于管理证书
* 2）创建 Issuer、Certificate 对象，让 cert-manager 签发证书并存放到 Secret
* 3）创建 rbac 并部署 Webhook, 挂载 2 中的 Secret 到容器中以开启 TLS
  * 可以修改启动命令中的 -kube-dns 和 -local-dns 参数来调整 KubeDNS 和 NodeLocalDNS 地址，默认为 10.96.0.10 和 169.254.20.10。
* 4）创建 WebhookConfig,增加 `cert-manager.io/inject-ca-from` annotation 用于自动注入 CA 证书


## 测试
给 Namespace 打上 label 然后创建 Pod，查看是否注入 DNSConfig.

```bash
kubectl create namespace myns
kubectl run busybox --image=busybox --restart=Never --namespace=myns --command -- sleep infinity
# 查看是否注入
kubectl -n myns get pod busybox -oyaml
# 测试能否正常解析
kubectl exec -it busybox -- nslookup nodelocaldns-webhook.kube-system.svc.cluster.local
```

到一个没有打 label 的 命名空间创建，应该不会注入了
```bash 
kubectl create namespace myns
kubectl run busybox --image=busybox --restart=Never --namespace=myns --command -- sleep infinity
kubectl -n myns get pod busybox -oyaml
```

