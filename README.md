# 日志收集
## 特性
  1. 追踪指定文件夹中文件的变化，同步到中心服务器
  2. 当中心服务器不可用时，客户端记录文件变化，待中心服务器可用后在同步到服务器

## Bug
  当文件超过6小时没有变更时，变更的第一条数据会丢弃

