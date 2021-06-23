# Kubernetes Webhook拦截请求
___
 
> 简单的webhook拦截代码demo


## validate 

> 拦截但不修改，只对请求的资源进行拦截，判断副本数量为少于2则不创建


## mutate

> 拦截并修改，对拦截的数据包进行修改，可以用来插入sidecar 容器