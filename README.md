# 高性能分布式即时通讯系统：基于 Kratos 的实现

## 1.项目实现

该即时通讯系统旨在提供以下核心功能：

- **实时消息传输**：支持用户之间的实时文本消息交流。
- **群组管理**：用户可以加入或创建聊天群组，实现多人聊天。
- **消息持久化**：所有消息将被存储到 MongoDB 中，支持历史消息查询。
- **高可用性和可扩展性**：通过分布式架构和负载均衡，确保系统的高可用性和可扩展性。
- **服务发现与负载均衡**：使用 Consul 和 Kong 实现动态服务发现与负载均衡。
- **消息分发**：利用 RocketMQ 作为分布式消息队列，实现高效的消息分发。
- **用户会话管理**：通过 Redis 存储用户会话信息，实现快速的用户信息查询和管理。
- **gRPC 接口**：提供多种 gRPC 接口，支持用户绑定、群组管理、消息发送与历史记录获取等操作。

## 2.目录构成

项目的目录结构设计合理，模块划分清晰，便于开发和维护。以下是项目的主要目录结构：

```
.
├── Dockerfile  
├── LICENSE
├── Makefile  
├── deploy // 环境快速搭建
├── README.md
├── api // 存放 proto 文件及生成的 Go 代码
│   └── chat
│       └── service
│           └── v1
│               ├── error_reason.pb.go
│               ├── error_reason.proto
│               ├── ws.pb.go
│               ├── ws.proto
│               └── ws_grpc.pb.go
├── app
│   └── chat
│       └── service
│           ├── cmd  // 项目启动入口
│           │   └── server
│           │       ├── main.go
│           │       ├── wire.go  // 使用 Wire 进行依赖注入
│           │       └── wire_gen.go
│           ├── configs  // 配置文件
│           │  ├── config.yaml
│           │  └── registry.yaml
│	    ├── internal  // 内部业务逻辑
│           │   ├── handler // WebSocket 处理器
│           │   ├── biz   // 业务逻辑层
│           │   │   ├── biz.go
│           │   │   └── ws.go
│           │   ├── conf  // 配置结构定义
│           │   │   ├── conf.pb.go
│           │   │   └── conf.proto
│           │   ├── model  // 数据库模型
│           │   ├── data  // 数据访问层
│           │   │   ├── data.go
│           │   │   └── ws.go
│           │   ├── server  // HTTP 和 gRPC 服务器配置
│           │   │   ├── grpc.go
│           │   │   ├── http.go
│           │   │   └── server.go
│           │   └── service  // 服务层实现
│           │       ├── greeter.go
│           │       └── service.go
│           ├── Dockerfile  
│           ├── Makefile  
│           └── utils // 工具函数
├── generate.go
├── go.mod
├── go.sum
└── third_party  // 第三方依赖的 proto 文件
    ├── README.md
    ├── google
    │   └── api
    │       ├── annotations.proto
    │       ├── http.proto
    │       └── httpbody.proto
    └── validate
        ├── README.md
        └── validate.proto
```

## 3.系统架构图
![系统架构图](https://blog.crdyjh.cn/wp-content/uploads/2024/10/wxjt-1.jpg)

## 4.本地运行

### Kratos
```
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
go mod tidy
kratos run
```
### 环境构建

在deploy目录下附带了快速构建环境的docker-compose,在每个牡蛎下分别执行`sudo docker-compose up -d`即可


