/*
Copyright 2018 Google LLC

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

package constants

const (
	// AllowAll is the value used to allow all images with CVEs, except for whitelisted CVEs
	AllowAll = "ALLOW_ALL"
	// BlockAll is the value used to block all images with CVEs, except for whitelisted CVEs
	BlockAll = "BLOCK_ALL"

	// InvalidImageSecPolicy is the key for labels and annotations
	InvalidImageSecPolicy           = "kritis.grafeas.io/invalidImageSecPolicy"
	InvalidImageSecPolicyLabelValue = "invalidImageSecPolicy"

	// ImageAttestation is the key for labels for indication attestaions.
	ImageAttestation             = "kritis.grafeas.io/attestation"
	NoAttestationsLabelValue     = "notAttested"
	PreviouslyAttestedLabelValue = "attested"

	// Breakglass is the key for the breakglass annotation
	Breakglass = "kritis.grafeas.io/breakglass"

	// A list of label values
	PreviouslyAttestedAnnotation = "Previously attested."
	NoAttestationsAnnotation     = "No valid attestations present. This pod will not be able to restart in future"

	// Atomic Container Signature type
	AtomicContainerSigType = "atomic container signature"

	// Public Key Private Key constants for Attestation Secrets.
	PrivateKey = "private"
	PublicKey  = "public"

	// Constants for Metadata Library
	PageSize          = int32(100)
	ResourceURLPrefix = "https://"
)

var (
	// GlobalImageWhitelist is a list of images that are globally whitelisted
	// They should always pass the webhook check
	GlobalImageWhitelist = []string{"gcr.io/kritis-project/kritis-server",
		"gcr.io/kritis-project/preinstall",
		"gcr.io/kritis-project/postinstall",
		"gcr.io/kritis-project/predelete"}
)
