# 基于请求属性的流量路由

## 概述

基于请求属性的流量路由功能允许您根据请求的特定属性（如请求头或查询参数）将流量路由到金丝雀部署。这在需要对特定用户群体或请求类型进行金丝雀发布时非常有用。

## 使用场景

1. **用户ID分片**: 根据用户ID将特定比例的用户流量路由到金丝雀版本
2. **客户端版本控制**: 根据客户端版本将特定客户端路由到金丝雀版本
3. **地理位置路由**: 根据地理位置信息将特定区域的用户路由到金丝雀版本

## 配置选项

基于请求属性的流量路由通过 `attributeRangeRouting` 字段进行配置：

```yaml
attributeRangeRouting:
  enabled: true
  headerName: "X-User-ID"        # 用于路由决策的请求头名称
  parameterName: "userId"        # 用于路由决策的查询参数名称（与headerName互斥）
  strategy: "consistent-hash"    # 路由策略（consistent-hash 或 range-based）
  initialPercentage: 10         # 初始路由到金丝雀的属性值百分比
  stepPercentage: 10            # 每次迭代增加的百分比
  maxPercentage: 50             # 最大路由到金丝雀的属性值百分比
  hashFunction: "fnv"           # 一致性哈希函数（fnv、md5、sha256）
  slotCount: 1000               # 一致性哈希的槽位数量
```

## 路由策略

### 一致性哈希 (consistent-hash)

一致性哈希策略使用哈希函数将请求属性值映射到固定数量的槽位中。这种方法确保相同的属性值始终路由到相同的目标（主版本或金丝雀版本），并且在调整流量比例时最小化重新路由的请求量。

### 范围基础 (range-based)

范围基础策略将属性值转换为数字，并根据配置的百分比范围决定路由目标。如果转换后的数字在指定范围内，则请求路由到金丝雀版本。

## 示例

以下示例展示了如何根据 `X-User-ID` 请求头的值进行流量路由：

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: podinfo
  namespace: test
spec:
  provider: istio
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: podinfo
  service:
    port: 9898
  attributeRangeRouting:
    enabled: true
    headerName: "X-User-ID"
    strategy: consistent-hash
    initialPercentage: 10
    stepPercentage: 10
    maxPercentage: 50
    hashFunction: fnv
    slotCount: 1000
  analysis:
    interval: 15s
    threshold: 15
    maxWeight: 50
    stepWeight: 10
```

在此示例中：
1. 初始时，具有特定 `X-User-ID` 值的 10% 用户将被路由到金丝雀版本
2. 每次迭代时，额外的 10% 用户将被路由到金丝雀版本
3. 最多 50% 的用户将被路由到金丝雀版本

## 注意事项

1. `headerName` 和 `parameterName` 是互斥的，只能指定其中一个
2. 当使用一致性哈希策略时，建议使用较大的 `slotCount` 值以获得更均匀的分布
3. 如果同时配置了基于权重的路由和基于属性的路由，基于属性的路由规则将优先匹配