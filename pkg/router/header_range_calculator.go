package router

import (
	"crypto/md5"
	"crypto/sha256"
	"hash/fnv"
	"strconv"
	"strings"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
)

// HeaderRangeCalculator provides methods to calculate if a request should be routed to canary
// based on request header values and consistent hashing or range-based strategies
type HeaderRangeCalculator struct{}

// NewHeaderRangeCalculator creates a new HeaderRangeCalculator
func NewHeaderRangeCalculator() *HeaderRangeCalculator {
	return &HeaderRangeCalculator{}
}

// ShouldRouteToCanary determines if a request should be routed to canary based on header value
// and the current percentage of traffic allocated to canary
func (hrc *HeaderRangeCalculator) ShouldRouteToCanary(
	routing *flaggerv1.CanaryAttributeRangeRouting,
	headerValue string,
	canaryPercentage int) bool {

	if !routing.Enabled || canaryPercentage <= 0 {
		return false
	}

	if routing.Strategy == "consistent-hash" {
		return hrc.consistentHashRouting(routing, headerValue, canaryPercentage)
	}

	// Default to range-based strategy
	return hrc.rangeBasedRouting(routing, headerValue, canaryPercentage)
}

// consistentHashRouting uses consistent hashing to determine if a request should go to canary
func (hrc *HeaderRangeCalculator) consistentHashRouting(
	routing *flaggerv1.CanaryAttributeRangeRouting,
	headerValue string,
	canaryPercentage int) bool {

	slotCount := 1000
	if routing.SlotCount > 0 {
		slotCount = routing.SlotCount
	}

	// Calculate hash of the header value
	hashValue := hrc.calculateHash(routing.HashFunction, headerValue)

	// Map to slot
	slot := hashValue % uint32(slotCount)

	// Calculate the number of slots that should go to canary
	canarySlots := (slotCount * canaryPercentage) / 100

	// Check if this slot is in the canary range
	return int(slot) < canarySlots
}

// rangeBasedRouting uses a simple range-based approach to determine if request should go to canary
func (hrc *HeaderRangeCalculator) rangeBasedRouting(
	routing *flaggerv1.CanaryAttributeRangeRouting,
	headerValue string,
	canaryPercentage int) bool {

	// For range-based routing, we convert the header value to a number
	// and check if it falls within the canary percentage range
	value := hrc.stringToNumber(headerValue)

	// Normalize to 0-100 range
	normalizedValue := value % 100

	return normalizedValue < canaryPercentage
}

// calculateHash calculates hash of a string using specified hash function
func (hrc *HeaderRangeCalculator) calculateHash(hashFunc, value string) uint32 {
	switch strings.ToLower(hashFunc) {
	case "md5":
		hash := md5.Sum([]byte(value))
		return uint32(hash[0])<<24 | uint32(hash[1])<<16 | uint32(hash[2])<<8 | uint32(hash[3])
	case "sha256":
		hash := sha256.Sum256([]byte(value))
		return uint32(hash[0])<<24 | uint32(hash[1])<<16 | uint32(hash[2])<<8 | uint32(hash[3])
	case "fnv":
		fallthrough
	default:
		h := fnv.New32a()
		h.Write([]byte(value))
		return h.Sum32()
	}
}

// stringToNumber converts a string to a number for range-based routing
func (hrc *HeaderRangeCalculator) stringToNumber(value string) int {
	// If the value is already a number, parse it
	if num, err := strconv.Atoi(value); err == nil {
		return num
	}

	// Otherwise, calculate a hash and use that
	hash := fnv.New32a()
	hash.Write([]byte(value))
	return int(hash.Sum32())
}

// GetCanaryPercentage calculates the current canary percentage based on the routing configuration
// and the current step in the canary process
func (hrc *HeaderRangeCalculator) GetCanaryPercentage(
	routing *flaggerv1.CanaryAttributeRangeRouting,
	currentStep int,
	maxWeight int) int {

	if !routing.Enabled {
		return 0
	}

	initialPercentage := 0
	if routing.InitialPercentage > 0 {
		initialPercentage = routing.InitialPercentage
	}

	stepPercentage := 10
	if routing.StepPercentage > 0 {
		stepPercentage = routing.StepPercentage
	}

	maxPercentage := 100
	if routing.MaxPercentage > 0 {
		maxPercentage = routing.MaxPercentage
	}

	// Calculate current percentage
	percentage := initialPercentage + (currentStep * stepPercentage)
	if percentage > maxPercentage {
		percentage = maxPercentage
	}

	return percentage
}
