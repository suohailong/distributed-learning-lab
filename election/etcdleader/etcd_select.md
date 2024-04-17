# etcd 选主

## 选举
```mermaid
    sequenceDiagram
    节点A ->> etcd: 开启事务 
    activate etcd
    etcd -->> 节点A: 回复 
    deactivate etcd
    节点A ->> etcd: 等待修订号小的键删除 
```