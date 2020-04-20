# GO RPC

### 项目描述
go-rpc 提供了公司内部 rpc 调用的 go 语言支持

### 功能列表
-  client 通过 jmtext 协议，连接本地 lark.
    - [sample](./example/jmtext-lark-client/sample.go)
    
-  client 通过 jmthrift 协议，直接连接远端 lark.
    - [sample](./example/jmthrift-client/sample.go)
    - 这种调用方式, client 这边是短连接
    
-  client 通过 jmthrift 协议，连接本地 lark.
    - [sample](./example/jmthrift-lark-client/sample.go)
    
-  server 通过 lark 提供 jmthrift 协议的服务. 
    - [sample](./example/jmthrift-lark-server/sample.go)
    - 注意 sample 中 owl context 的使用方法
    - 默认会开启 rpcmon 的数据生成和上报，如果想自行设置，请查看 `WithRpcmonStats` 和 `WithRpcmonEnable` 函数
    
-  server 通过 lark 提供 jmtext 协议的服务.
    - [sample](./example/jmtext-lark-server/sample.go)
    
### 注意事项
-  使用 thrift 命令处理 IDL 文件的时候， 建议使用 0.10.0 版本，或者附近的小版本. 版本差别太大可能存在不兼容的情况。
   参考命令 ``` thrift -r --gen go xx路径.thrift```。
   macOS 可以使用本项目 bin 路径下的 thrift-darwin 程序
-  thrift package 的版本请参考本项目 [go.mod](./go.mod) 中的版本，否则可能存在不兼容情况

### 时间线
- 2019/12/17 v0.1 
    - 完成 rpc 的基本调用功能
- 2019/12/20 v0.2 
    - 支持业务使用 owl context 相关功能
- 2020/1/3 v0.3
    - rpcmon 功能添加完成
   
### TODO
- client 连接 lark 的时候，提供连接池
- client 通过 jmtext 协议，直接连接远端 lark.
