/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package canary

import (
	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/fluxcd/flagger/pkg/utils"
)

// hasSpecChanged computes the hash of the spec and compares it with the
// last applied spec, if the last applied hash is different but not equal
// to last promoted one the it returns true
func hasSpecChanged(cd *flaggerv1.Canary, spec interface{}) (bool, error) {
	if cd.Status.LastAppliedSpec == "" {
		return true, nil
	}

	newHash := utils.ComputeHash(spec)

	// do not trigger a canary deployment on manual rollback
	if cd.Status.LastPromotedSpec == newHash {
		return false, nil
	}

	if cd.Status.LastAppliedSpec != newHash {
		return true, nil
	}

	return false, nil
}
