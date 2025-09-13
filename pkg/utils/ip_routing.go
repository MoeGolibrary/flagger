package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"net"
	"strings"
)

type IPRangeCalculator struct {
	Strategy     string
	HashFunction string
	SlotCount    int
}

func NewIPRangeCalculator(strategy, hashFunction string, slotCount int) *IPRangeCalculator {
	if slotCount == 0 {
		slotCount = 1000 // default slot count
	}
	if hashFunction == "" {
		hashFunction = "fnv" // default hash function
	}
	return &IPRangeCalculator{
		Strategy:     strategy,
		HashFunction: hashFunction,
		SlotCount:    slotCount,
	}
}

func (calc *IPRangeCalculator) CalculateIPRanges(percentage int) ([]string, error) {
	if percentage < 0 || percentage > 100 {
		return nil, fmt.Errorf("percentage must be between 0 and 100")
	}

	switch calc.Strategy {
	case "consistent-hash":
		return calc.calculateConsistentHashRanges(percentage)
	case "range-based":
		return calc.calculateRangeBasedRanges(percentage)
	default:
		return calc.calculateConsistentHashRanges(percentage)
	}
}

func (calc *IPRangeCalculator) calculateConsistentHashRanges(percentage int) ([]string, error) {
	targetSlots := (calc.SlotCount * percentage) / 100
	if targetSlots == 0 && percentage > 0 {
		targetSlots = 1
	}

	var ranges []string
	for i := 0; i < targetSlots; i++ {
		ranges = append(ranges, fmt.Sprintf("slot-%d", i))
	}

	return ranges, nil
}

func (calc *IPRangeCalculator) calculateRangeBasedRanges(percentage int) ([]string, error) {
	var ranges []string
	
	subnetsNeeded := (256 * percentage) / 100
	if subnetsNeeded == 0 && percentage > 0 {
		subnetsNeeded = 1
	}

	for i := 0; i < subnetsNeeded && i < 256; i++ {
		ranges = append(ranges, fmt.Sprintf("10.0.%d.0/24", i))
	}

	return ranges, nil
}

func (calc *IPRangeCalculator) HashIP(ip string) (int, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ip)
	}

	var hashValue uint64
	switch calc.HashFunction {
	case "fnv":
		hasher := fnv.New64a()
		hasher.Write(parsedIP)
		hashValue = hasher.Sum64()
	case "md5":
		hasher := md5.New()
		hasher.Write(parsedIP)
		hashBytes := hasher.Sum(nil)
		hashValue = uint64(hashBytes[0])<<56 | uint64(hashBytes[1])<<48 | uint64(hashBytes[2])<<40 | uint64(hashBytes[3])<<32 |
			uint64(hashBytes[4])<<24 | uint64(hashBytes[5])<<16 | uint64(hashBytes[6])<<8 | uint64(hashBytes[7])
	case "sha256":
		hasher := sha256.New()
		hasher.Write(parsedIP)
		hashBytes := hasher.Sum(nil)
		hashValue = uint64(hashBytes[0])<<56 | uint64(hashBytes[1])<<48 | uint64(hashBytes[2])<<40 | uint64(hashBytes[3])<<32 |
			uint64(hashBytes[4])<<24 | uint64(hashBytes[5])<<16 | uint64(hashBytes[6])<<8 | uint64(hashBytes[7])
	default:
		hasher := fnv.New64a()
		hasher.Write(parsedIP)
		hashValue = hasher.Sum64()
	}

	return int(hashValue % uint64(calc.SlotCount)), nil
}

func (calc *IPRangeCalculator) ShouldRouteToCanary(ip string, percentage int) (bool, error) {
	if percentage == 0 {
		return false, nil
	}
	if percentage >= 100 {
		return true, nil
	}

	switch calc.Strategy {
	case "consistent-hash":
		slot, err := calc.HashIP(ip)
		if err != nil {
			return false, err
		}
		targetSlots := (calc.SlotCount * percentage) / 100
		return slot < targetSlots, nil
	case "range-based":
		return calc.isIPInRanges(ip, percentage)
	default:
		return calc.ShouldRouteToCanary(ip, percentage)
	}
}

func (calc *IPRangeCalculator) isIPInRanges(ip string, percentage int) (bool, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, fmt.Errorf("invalid IP address: %s", ip)
	}

	ranges, err := calc.calculateRangeBasedRanges(percentage)
	if err != nil {
		return false, err
	}

	for _, cidr := range ranges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true, nil
		}
	}

	return false, nil
}

func (calc *IPRangeCalculator) GenerateHashRangeRegex(percentage int) (string, error) {
	if percentage == 0 {
		return "^$", nil // Match nothing
	}
	if percentage >= 100 {
		return ".*", nil // Match everything
	}

	targetSlots := (calc.SlotCount * percentage) / 100
	if targetSlots == 0 {
		targetSlots = 1
	}

	var patterns []string
	for i := 0; i < targetSlots; i++ {
		patterns = append(patterns, fmt.Sprintf("slot-%d", i))
	}

	return fmt.Sprintf("^(%s)$", strings.Join(patterns, "|")), nil
}
