/*
 * Copyright 2026 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package keda

import "knative.dev/eventing-kafka-broker/control-plane/pkg/flags"

// init drops the global "kubeconfig" flag that
// sigs.k8s.io/controller-runtime/pkg/client/config registers in its own init().
//
// This package imports the KEDA apis (github.com/kedacore/keda/v2/apis/keda/v1alpha1),
// which — on KEDA's kubebuilder v3 layout — carry admission-webhook code that
// imports controller-runtime. That controller-runtime init() registers a global
// "kubeconfig" flag; knative's sharedmain (in controllers) and the e2e test
// harness later register the same flag, panicking with "flag redefined: kubeconfig".
//
// Because this package depends on controller-runtime, Go guarantees
// controller-runtime's init() runs before this one, so the flag is already
// registered by the time we drop it, and no other init() re-adds it before the
// runtime registration. Doing this here — rather than in each main() — covers
// every binary that touches KEDA, including go-test-generated test binaries that
// pull the KEDA apis in transitively (e.g. via test/rekt features).
func init() {
	flags.DropControllerRuntimeKubeconfigFlag()
}
