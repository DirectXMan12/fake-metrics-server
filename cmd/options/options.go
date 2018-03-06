// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

import (
	"github.com/spf13/pflag"

	genericoptions "k8s.io/apiserver/pkg/server/options"
)

type FakeMetricsServerRunOptions struct {
	// genericoptions.ReccomendedOptions - EtcdOptions
	SecureServing  *genericoptions.SecureServingOptions
	Authentication *genericoptions.DelegatingAuthenticationOptions
	Authorization  *genericoptions.DelegatingAuthorizationOptions
	Features       *genericoptions.FeatureOptions

	// Only to be used to for testing
	DisableAuthForTesting bool
}

func NewFakeMetricsServerRunOptions() *FakeMetricsServerRunOptions {
	return &FakeMetricsServerRunOptions{
		SecureServing:  genericoptions.NewSecureServingOptions(),
		Authentication: genericoptions.NewDelegatingAuthenticationOptions(),
		Authorization:  genericoptions.NewDelegatingAuthorizationOptions(),
		Features:       genericoptions.NewFeatureOptions(),
	}
}

func (h *FakeMetricsServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	h.SecureServing.AddFlags(fs)
	h.Authentication.AddFlags(fs)
	h.Authorization.AddFlags(fs)
	h.Features.AddFlags(fs)
}
