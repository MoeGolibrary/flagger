package utils

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"hash/fnv"
	"k8s.io/apimachinery/pkg/util/rand"
)

// ComputeHash returns a hash value calculated from a spec using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func ComputeHash(spec interface{}) string {
	hasher := fnv.New64a()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", spec)

	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum64()))
}
