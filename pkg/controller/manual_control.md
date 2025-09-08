# 人工介入流程功能说明

## 概述

Flagger 提供了人工介入流程功能，允许用户在 Canary 部署过程中暂停流量转移，并在需要时手动恢复自动化流程。此功能特别适用于需要在特定流量比例下进行验证的场景，例如在高峰期验证新版本的稳定性。

## 功能特性

1. **流量比例暂停**：可以在指定流量权重处暂停流量转移过程
2. **手动恢复**：支持手动恢复自动化流量转移过程
3. **状态管理**：通过 `CanaryPhaseWaiting` 状态标识暂停状态
4. **持久化配置**：通过 Canary CRD 中的 `ManualStep` 字段配置

## 数据结构

### CanaryManualStep

```go
type CanaryManualStep struct {
    // Weight 定义了在手动模式下路由到 canary 的流量权重
    Weight int `json:"weight,omitempty"`

    // Resume 指示 canary 是否应该恢复自动流量转移
    Resume bool `json:"resume,omitempty"`
}
```

### 使用示例

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: podinfo
spec:
  # 其他配置...
  manualStep:
    weight: 10
  # 或者恢复自动流程
  # manualStep:
  #   weight: 10
  #   resume: true
```

## 实现原理

### 暂停逻辑

当满足以下条件时，控制器会暂停流量转移：

1. `ManualStep` 字段被定义
2. `ManualStep.Resume` 为 `false`
3. 当前 canary 权重达到或超过 `ManualStep.Weight`

暂停后，控制器会：

1. 将 Canary 状态设置为 `CanaryPhaseWaiting`
2. 保持流量在指定权重
3. 停止进一步的自动化流量转移

### 恢复逻辑

当 `ManualStep.Resume` 设置为 `true` 时：

1. 控制器会重置 `ManualStep` 字段
2. 继续自动流量转移过程
3. 将 Canary 状态从 `CanaryPhaseWaiting` 恢复为 `CanaryPhaseProgressing`

## 代码实现

### 关键代码位置

- [scheduler.go](file:///Users/hanyunpeng/Projects/flagger/pkg/controller/scheduler.go) - 主要调度逻辑
- [scheduler_hooks.go](file:///Users/hanyunpeng/Projects/flagger/pkg/controller/scheduler_hooks.go) - Webhook 相关逻辑

### 核心处理逻辑

```go
// 检查手动步骤
if cd.Spec.ManualStep != nil {
    if !cd.Spec.ManualStep.Resume {
        // 手动步骤已定义但未恢复
        if canaryWeight >= cd.Spec.ManualStep.Weight {
            // 已达到手动步骤权重，暂停
            c.recordEventInfof(cd, "Canary is paused at %d%% traffic weight. Waiting for manual resume.", 
                cd.Spec.ManualStep.Weight)
            
            // 设置状态为等待
            if cd.Status.Phase != flaggerv1.CanaryPhaseWaiting {
                if err := canaryController.SetStatusPhase(cd, flaggerv1.CanaryPhaseWaiting); err != nil {
                    c.recordEventWarningf(cd, "%v", err)
                }
            }
            
            // 保持指定权重的流量
            if err := meshRouter.SetRoutes(cd, 100-cd.Spec.ManualStep.Weight, cd.Spec.ManualStep.Weight, false); err != nil {
                c.recordEventWarningf(cd, "%v", err)
            }
            c.recorder.SetWeight(cd, 100-cd.Spec.ManualStep.Weight, cd.Spec.ManualStep.Weight)
            return
        }
    } else {
        // 手动步骤已恢复，继续自动流量转移
        // 重置手动步骤
        defer func() {
            cd.Spec.ManualStep = nil
            _, err := c.flaggerClient.FlaggerV1beta1().Canaries(cd.Namespace).Update(context.TODO(), cd, metav1.UpdateOptions{})
            if err != nil {
                c.recordEventWarningf(cd, "Failed to reset manual step: %v", err)
            }
        }()
        
        c.recordEventInfof(cd, "Resuming automated traffic shifting from %d%% canary weight", canaryWeight)
    }
}
```

## 使用场景

### 长时间流量灰度

适用于需要在特定流量比例下长时间观察系统表现的场景：

1. 配置 10% 流量到 Canary 版本
2. 在高峰期观察系统表现
3. 验证稳定后手动恢复自动化流程

### 手动验证

在自动化流程中插入人工验证步骤：

1. 在关键流量节点暂停
2. 执行人工验证
3. 确认无误后恢复自动化流程

## 注意事项

1. 暂停状态下，流量会持续保持在指定权重
2. 恢复自动化流程后，`ManualStep` 字段会被自动清除
3. 可以通过修改 Canary 资源来随时调整或取消手动步骤
4. 暂停期间仍然会执行指标检查和 Webhook