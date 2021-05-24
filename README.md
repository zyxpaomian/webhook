# go-http-frame
___
 
自用golang的simple-http 框架，restful api， 支持token认证


### 安装
``` shell
go env -w GOPROXY=https://goproxy.cn,https://goproxy.io,direct
go env -w GO111MODULE=on

# 运行
make run 

# 编译
make compile
```
 
### 主要目录结构
* common 基本库，如日志，mysql 驱动, 字符转换，配置加载等
* controller 控制层，控制agent基本的逻辑代码位置，如listagent, delagent等
* dao dao层，数据层
* http mux+http封装的http 层，接口封装
* structs 统一的结构体位置
