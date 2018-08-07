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

package container

import (
	"encoding/json"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/grafeas/kritis/pkg/kritis/attestation"
	"github.com/grafeas/kritis/pkg/kritis/constants"
	"github.com/grafeas/kritis/pkg/kritis/secrets"
)

// AtomicSig represents Red Hat’s Atomic Host attestation signature format
// defined here https://github.com/aweiteka/image/blob/e5a20d98fe698732df2b142846d007b45873627f/docs/signature.md
type AtomicSig struct {
	Critical *Critical         `json:"critical"`
	Optional map[string]string `json:"optional,omitempty"`
}

func NewAtomicSig(image string, optional map[string]string) (*AtomicSig, error) {
	critical, err := NewCritical(image)
	if err != nil {
		return nil, err
	}
	return &AtomicSig{
		Critical: critical,
		Optional: optional,
	}, nil
}

type Critical struct {
	Identity *Identity `json:"identity"`
	Image    *Image    `json:"image"`
	Type     string    `json:"type"`
}

func NewCritical(image string) (*Critical, error) {
	digest, err := name.NewDigest(image, name.StrictValidation)
	if err != nil {
		return nil, err
	}
	return &Critical{
		Identity: NewIdentity(digest.Repository.Name()),
		Image:    NewImage(digest.DigestStr()),
		Type:     constants.AtomicContainerSigType,
	}, nil
}

type Identity struct {
	DockerRef string `json:"docker-reference"`
}

func NewIdentity(image string) *Identity {
	return &Identity{
		DockerRef: image,
	}
}

type Image struct {
	DockerDigest string `json:"docker-manifest-digest"`
}

func NewImage(digest string) *Image {
	return &Image{
		DockerDigest: digest,
	}
}

func (acs *AtomicSig) JSON() (string, error) {
	bytes, err := json.Marshal(acs)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (acs *AtomicSig) CreateAttestationSignature(pgpSigningKey *secrets.PGPSigningSecret) (string, error) {
	hostStr, err := acs.JSON()
	if err != nil {
		return "", err
	}
	return attestation.CreateMessageAttestation(pgpSigningKey.PublicKey, pgpSigningKey.PrivateKey, hostStr)
}

func (acs *AtomicSig) VerifyAttestationSignature(publicKey string, attestationHash string) error {
	hostStr, err := acs.JSON()
	if err != nil {
		return err
	}
	return attestation.VerifyMessageAttestation(publicKey, attestationHash, hostStr)
}
