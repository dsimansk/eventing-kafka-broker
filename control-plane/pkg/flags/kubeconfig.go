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

package flags

import "flag"

// DropControllerRuntimeKubeconfigFlag removes the global "kubeconfig" flag that
// sigs.k8s.io/controller-runtime/pkg/client/config registers in its init().
//
// The KEDA apis package (github.com/kedacore/keda/v2/apis/keda/v1alpha1) ships
// admission-webhook validators that import controller-runtime, so every binary
// importing the KEDA client transitively runs that init() and registers
// "kubeconfig" on flag.CommandLine. knative's sharedmain later registers the same
// flag unconditionally (environment.ClientConfig.InitFlags), which panics with
// "flag redefined: kubeconfig". Dropping controller-runtime's registration here,
// before sharedmain runs, lets knative own the flag.
//
// Call this at the very start of main(), before sharedmain.MainNamed. It is a
// no-op if the flag is not registered.
func DropControllerRuntimeKubeconfigFlag() {
	if flag.CommandLine.Lookup("kubeconfig") == nil {
		return
	}

	// A flag cannot be removed from a FlagSet, so rebuild flag.CommandLine while
	// preserving every flag except "kubeconfig".
	pruned := flag.NewFlagSet(flag.CommandLine.Name(), flag.ExitOnError)
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name == "kubeconfig" {
			return
		}
		pruned.Var(f.Value, f.Name, f.Usage)
	})
	flag.CommandLine = pruned
}
