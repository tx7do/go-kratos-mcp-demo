# 基于 Go-Kratos 与 MCP 的模块化推荐服务

简短说明：本项目演示如何用 Go-Kratos 框架结合 MCP（模块化协同协议）构建可扩展、可观测的推荐服务，包含召回 / 过滤 / 排序等模块化流程与测试范例。

## 核心特性

- 标准化契约：以 Protobuf 定义服务与上下文，便于多语言客户端与版本管理。
- 模块化协同：通过 MCP 将召回、排序、过滤等模块串联，模块可独立开发与部署。
- 多协议支持：同服务同时暴露 MCP 与 HTTP 接口，满足不同调用场景。
- 可观测性与测试：支持链路追踪、日志、指标，并提供端到端与单元测试示例（兼容 Protobuf JSON 中 int64 被编码为 string 的情况）。

## 架构概览

- 请求流：客户端 → MCP 请求 → 解析为内部 Context → 召回 → 排序 → 过滤 → 封装为 CallToolResult 返回。
- 依赖注入：使用 Kratos 的 wire 管理模块依赖，s/recall/rank/filter 服务独立实现逻辑并通过接口解耦。
- 工具注册：在 MCP 服务启动时注册 recommend 工具并声明输入 Schema，调用方按协议传参即可。

## 快速上手

1. 生成代码：`make api openapi`
2. 生成依赖注入：`make wire`
3. 启动服务：`make run`（同时启动 MCP 与 HTTP 服务）
4. 本地测试：使用 MCP 客户端模拟调用，测试需注意对 int64 字段在 JSON 中可能为字符串的兼容断言。

## 测试与注意点

- 推荐在测试中对 responseTime 使用 time.RFC3339Nano 校验；对 historyItemIds、blacklistItemIds 等 ID 列表做兼容字符串/数字的断言。
- 在断言数值时，建议编写通用转换函数以支持 float64/string/int64 三种表现形式。

## 结论

本项目示例展示了「框架 + 协议」的协同价值：Kratos 提供服务治理与依赖注入能力，MCP 提供模块化协同标准。对需要模块化、可扩展且具备跨服务契约的推荐系统，该方案适合生产化落地。

## 项目链接

- GitHub项目地址: <https://github.com/tx7do/go-kratos-mcp-demo>
- Gitee项目地址: <https://gitee.com/tx7do/go-kratos-mcp-demo>
- MCP 协议封装库: <https://github.com/mark3labs/mcp-go>
- Kratos Transport MCP 扩展: [github.com/tx7do/kratos-transport/transport/mcp](https://github.com/tx7do/kratos-transport/tree/main/transport/mcp)
