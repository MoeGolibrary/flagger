package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIPRangeCalculator(t *testing.T) {
	calc := NewIPRangeCalculator("consistent-hash", "fnv", 1000)
	assert.Equal(t, "consistent-hash", calc.Strategy)
	assert.Equal(t, "fnv", calc.HashFunction)
	assert.Equal(t, 1000, calc.SlotCount)

	calc2 := NewIPRangeCalculator("", "", 0)
	assert.Equal(t, "fnv", calc2.HashFunction)
	assert.Equal(t, 1000, calc2.SlotCount)
}

func TestHashIP(t *testing.T) {
	calc := NewIPRangeCalculator("consistent-hash", "fnv", 1000)

	hash, err := calc.HashIP("192.168.1.1")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, hash, 0)
	assert.Less(t, hash, 1000)

	_, err = calc.HashIP("invalid-ip")
	assert.Error(t, err)

	hash1, _ := calc.HashIP("192.168.1.1")
	hash2, _ := calc.HashIP("192.168.1.1")
	assert.Equal(t, hash1, hash2)

	hash3, _ := calc.HashIP("192.168.1.2")
	assert.NotEqual(t, hash1, hash3)
}

func TestShouldRouteToCanary(t *testing.T) {
	calc := NewIPRangeCalculator("consistent-hash", "fnv", 1000)

	shouldRoute, err := calc.ShouldRouteToCanary("192.168.1.1", 0)
	require.NoError(t, err)
	assert.False(t, shouldRoute)

	shouldRoute, err = calc.ShouldRouteToCanary("192.168.1.1", 100)
	require.NoError(t, err)
	assert.True(t, shouldRoute)

	shouldRoute1, err := calc.ShouldRouteToCanary("192.168.1.1", 50)
	require.NoError(t, err)
	shouldRoute2, err := calc.ShouldRouteToCanary("192.168.1.1", 50)
	require.NoError(t, err)
	assert.Equal(t, shouldRoute1, shouldRoute2)
}

func TestGenerateHashRangeRegex(t *testing.T) {
	calc := NewIPRangeCalculator("consistent-hash", "fnv", 1000)

	regex, err := calc.GenerateHashRangeRegex(0)
	require.NoError(t, err)
	assert.Equal(t, "^$", regex)

	regex, err = calc.GenerateHashRangeRegex(100)
	require.NoError(t, err)
	assert.Equal(t, ".*", regex)

	regex, err = calc.GenerateHashRangeRegex(10)
	require.NoError(t, err)
	assert.Contains(t, regex, "slot-0")
	assert.Contains(t, regex, "slot-99")
	assert.NotContains(t, regex, "slot-100")
}

func TestCalculateIPRanges(t *testing.T) {
	calc := NewIPRangeCalculator("consistent-hash", "fnv", 1000)

	ranges, err := calc.CalculateIPRanges(10)
	require.NoError(t, err)
	assert.Len(t, ranges, 100) // 10% of 1000 slots

	_, err = calc.CalculateIPRanges(-1)
	assert.Error(t, err)

	_, err = calc.CalculateIPRanges(101)
	assert.Error(t, err)
}

func TestHashIPDifferentFunctions(t *testing.T) {
	ip := "192.168.1.1"

	calcFNV := NewIPRangeCalculator("consistent-hash", "fnv", 1000)
	calcMD5 := NewIPRangeCalculator("consistent-hash", "md5", 1000)
	calcSHA256 := NewIPRangeCalculator("consistent-hash", "sha256", 1000)

	hashFNV, err := calcFNV.HashIP(ip)
	require.NoError(t, err)

	hashMD5, err := calcMD5.HashIP(ip)
	require.NoError(t, err)

	hashSHA256, err := calcSHA256.HashIP(ip)
	require.NoError(t, err)

	assert.NotEqual(t, hashFNV, hashMD5)
	assert.NotEqual(t, hashFNV, hashSHA256)
	assert.NotEqual(t, hashMD5, hashSHA256)

	assert.GreaterOrEqual(t, hashFNV, 0)
	assert.Less(t, hashFNV, 1000)
	assert.GreaterOrEqual(t, hashMD5, 0)
	assert.Less(t, hashMD5, 1000)
	assert.GreaterOrEqual(t, hashSHA256, 0)
	assert.Less(t, hashSHA256, 1000)
}

func TestRangeBasedStrategy(t *testing.T) {
	calc := NewIPRangeCalculator("range-based", "fnv", 1000)

	ranges, err := calc.CalculateIPRanges(10)
	require.NoError(t, err)
	assert.Greater(t, len(ranges), 0)

	for _, r := range ranges {
		assert.Contains(t, r, "/24")
		assert.Contains(t, r, "10.0.")
	}
}

func TestIsIPInRanges(t *testing.T) {
	calc := NewIPRangeCalculator("range-based", "fnv", 1000)

	inRange, err := calc.isIPInRanges("10.0.0.1", 10)
	require.NoError(t, err)
	assert.True(t, inRange)

	inRange, err = calc.isIPInRanges("192.168.1.1", 10)
	require.NoError(t, err)
	assert.False(t, inRange)

	_, err = calc.isIPInRanges("invalid-ip", 10)
	assert.Error(t, err)
}
