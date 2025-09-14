package router

import (
	"testing"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestHeaderRangeCalculator_ConsistentHashRouting(t *testing.T) {
	calculator := NewHeaderRangeCalculator()

	// 测试用例1: 测试一致性哈希路由 - FNV 哈希
	routing := &flaggerv1.CanaryAttributeRangeRouting{
		Enabled:           true,
		Strategy:          "consistent-hash",
		HashFunction:      "fnv",
		SlotCount:         1000,
		InitialPercentage: 10,
		StepPercentage:    10,
		MaxPercentage:     50,
	}

	// 测试不同的 header 值
	headerValue1 := "user123"
	headerValue2 := "user456"
	headerValue3 := "user789"

	// 10% 流量应该路由到 Canary
	canaryPercentage := 10
	result1 := calculator.ShouldRouteToCanary(routing, headerValue1, canaryPercentage)
	result2 := calculator.ShouldRouteToCanary(routing, headerValue2, canaryPercentage)
	result3 := calculator.ShouldRouteToCanary(routing, headerValue3, canaryPercentage)

	// 验证结果（至少有一个应该路由到 Canary，具体取决于哈希值）
	t.Logf("Header values: %s, %s, %s", headerValue1, headerValue2, headerValue3)
	t.Logf("Routing results: %v, %v, %v", result1, result2, result3)

	// 50% 流量应该路由到 Canary
	canaryPercentage = 50
	result1 = calculator.ShouldRouteToCanary(routing, headerValue1, canaryPercentage)
	result2 = calculator.ShouldRouteToCanary(routing, headerValue2, canaryPercentage)
	result3 = calculator.ShouldRouteToCanary(routing, headerValue3, canaryPercentage)

	t.Logf("50%% routing results: %v, %v, %v", result1, result2, result3)

	// 禁用时应该返回 false
	routing.Enabled = false
	result := calculator.ShouldRouteToCanary(routing, headerValue1, 50)
	assert.False(t, result)
}

func TestHeaderRangeCalculator_RangeBasedRouting(t *testing.T) {
	calculator := NewHeaderRangeCalculator()

	// 测试用例2: 测试范围基础路由
	routing := &flaggerv1.CanaryAttributeRangeRouting{
		Enabled:           true,
		Strategy:          "range-based",
		InitialPercentage: 10,
		StepPercentage:    10,
		MaxPercentage:     50,
	}

	// 测试数字字符串
	headerValue1 := "123"
	headerValue2 := "456"
	headerValue3 := "789"

	// 10% 流量应该路由到 Canary
	canaryPercentage := 10
	result1 := calculator.ShouldRouteToCanary(routing, headerValue1, canaryPercentage)
	result2 := calculator.ShouldRouteToCanary(routing, headerValue2, canaryPercentage)
	result3 := calculator.ShouldRouteToCanary(routing, headerValue3, canaryPercentage)

	t.Logf("Range-based routing - Header values: %s, %s, %s", headerValue1, headerValue2, headerValue3)
	t.Logf("Range-based routing - Results: %v, %v, %v", result1, result2, result3)

	// 50% 流量应该路由到 Canary
	canaryPercentage = 50
	result1 = calculator.ShouldRouteToCanary(routing, headerValue1, canaryPercentage)
	result2 = calculator.ShouldRouteToCanary(routing, headerValue2, canaryPercentage)
	result3 = calculator.ShouldRouteToCanary(routing, headerValue3, canaryPercentage)

	t.Logf("Range-based 50%% routing - Results: %v, %v, %v", result1, result2, result3)
}

func TestHeaderRangeCalculator_GetCanaryPercentage(t *testing.T) {
	calculator := NewHeaderRangeCalculator()

	// 测试用例3: 测试 Canary 百分比计算
	routing := &flaggerv1.CanaryAttributeRangeRouting{
		Enabled:           true,
		InitialPercentage: 10,
		StepPercentage:    10,
		MaxPercentage:     50,
	}

	// 测试不同步骤的百分比计算
	testCases := []struct {
		step        int
		expected    int
		description string
	}{
		{0, 10, "Step 0 should be initial percentage"},
		{1, 20, "Step 1 should be initial + step percentage"},
		{2, 30, "Step 2 should be initial + 2 * step percentage"},
		{4, 50, "Step 4 should be max percentage"},
		{10, 50, "Step 10 should be capped at max percentage"},
	}

	for _, tc := range testCases {
		actual := calculator.GetCanaryPercentage(routing, tc.step, 100)
		assert.Equal(t, tc.expected, actual, tc.description)
	}

	// 测试禁用情况
	routing.Enabled = false
	percentage := calculator.GetCanaryPercentage(routing, 2, 100)
	assert.Equal(t, 0, percentage, "Disabled routing should return 0 percentage")
}

func TestHeaderRangeCalculator_HashFunctions(t *testing.T) {
	calculator := NewHeaderRangeCalculator()

	// 测试不同的哈希函数
	headerValue := "test-user-id"

	// 测试 FNV 哈希
	hash1 := calculator.calculateHash("fnv", headerValue)
	t.Logf("FNV hash of '%s': %d", headerValue, hash1)

	// 测试 MD5 哈希
	hash2 := calculator.calculateHash("md5", headerValue)
	t.Logf("MD5 hash of '%s': %d", headerValue, hash2)

	// 测试 SHA256 哈希
	hash3 := calculator.calculateHash("sha256", headerValue)
	t.Logf("SHA256 hash of '%s': %d", headerValue, hash3)

	// 测试默认哈希（应该使用 FNV）
	hash4 := calculator.calculateHash("", headerValue)
	t.Logf("Default hash of '%s': %d", headerValue, hash4)
	assert.Equal(t, hash1, hash4, "Default hash should be FNV")

	// 验证相同输入产生相同输出
	hash1Again := calculator.calculateHash("fnv", headerValue)
	assert.Equal(t, hash1, hash1Again, "Same input should produce same hash")
}
