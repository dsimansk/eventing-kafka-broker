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

import (
	"flag"
	"testing"
)

// withFreshCommandLine swaps flag.CommandLine for the duration of a test and
// restores it afterwards, so tests don't clobber the global flag set.
func withFreshCommandLine(t *testing.T) *flag.FlagSet {
	t.Helper()
	orig := flag.CommandLine
	t.Cleanup(func() { flag.CommandLine = orig })
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	return flag.CommandLine
}

func TestDropControllerRuntimeKubeconfigFlag(t *testing.T) {
	withFreshCommandLine(t)

	// Simulate controller-runtime's init(): register "kubeconfig", plus an
	// unrelated flag we must preserve.
	var kubeconfig, other string
	flag.CommandLine.StringVar(&kubeconfig, "kubeconfig", "cr-default", "controller-runtime")
	flag.CommandLine.StringVar(&other, "keep-me", "v", "unrelated")

	DropControllerRuntimeKubeconfigFlag()

	if flag.CommandLine.Lookup("kubeconfig") != nil {
		t.Fatal("kubeconfig flag should have been removed")
	}
	if flag.CommandLine.Lookup("keep-me") == nil {
		t.Fatal("unrelated flag keep-me must be preserved")
	}

	// knative sharedmain re-registers "kubeconfig" unconditionally; this must
	// no longer panic now that the earlier registration is gone.
	var knativeKubeconfig string
	flag.CommandLine.StringVar(&knativeKubeconfig, "kubeconfig", "knative-default", "knative")
	if got := flag.CommandLine.Lookup("kubeconfig"); got == nil || got.DefValue != "knative-default" {
		t.Fatalf("expected knative kubeconfig registration to win, got %+v", got)
	}
}

func TestDropControllerRuntimeKubeconfigFlag_NoOpWhenAbsent(t *testing.T) {
	fs := withFreshCommandLine(t)

	var other string
	flag.CommandLine.StringVar(&other, "keep-me", "v", "unrelated")

	DropControllerRuntimeKubeconfigFlag()

	if flag.CommandLine != fs {
		t.Fatal("flag.CommandLine should be left untouched when kubeconfig is absent")
	}
	if flag.CommandLine.Lookup("keep-me") == nil {
		t.Fatal("unrelated flag keep-me must be preserved")
	}
}
