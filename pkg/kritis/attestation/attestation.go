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

// Package attestation defines methods to attest a message using Pgp Private and
// Public Key pair.
package attestation

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/grafeas/kritis/pkg/kritis/admission/constants"
	"github.com/pkg/errors"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/openpgp/s2k"
)

var pgpConfig = packet.Config{
	// Use Sha256
	DefaultHash:            crypto.SHA256,
	DefaultCipher:          packet.CipherAES256,
	DefaultCompressionAlgo: packet.CompressionZLIB,
	CompressionConfig: &packet.CompressionConfig{
		Level: packet.DefaultCompression,
	},
	RSABits: constants.RSABits,
}

// VerifyMessageAttestation verifies if the image is attested using the Base64
// encoded public key.
func VerifyMessageAttestation(pubKeyEnc string, attestationHash string, message string) error {
	pemPublicKey, err := base64.StdEncoding.DecodeString(pubKeyEnc)
	if err != nil {
		return err
	}

	attestation, err := base64.StdEncoding.DecodeString(attestationHash)
	if err != nil {
		return err
	}

	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(pemPublicKey))
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(attestation)
	armorBlock, err := armor.Decode(buf)
	if err != nil {
		return errors.Wrap(err, "could not decode armor signature")
	}
	md, err := openpgp.ReadMessage(armorBlock.Body, keyring, nil, &pgpConfig)
	if err != nil {
		return errors.Wrap(err, "could not read armor signature")
	}

	// MessageDetails.UnverifiedBody signature is not verified until we read it.
	// This will call PublicKey.VerifySignature for the keys in the keyring.
	plaintext, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return errors.Wrap(err, "could not verify armor signature")
	}
	// Make sure after reading the UnverifiedBody above, there is no signature error.
	if md.SignatureError != nil || md.Signature == nil {
		return fmt.Errorf("bad signature found: %s or no signature found for given key", md.SignatureError)
	}

	// Finally, make sure the signature is over the right message.
	if string(plaintext) != message {
		return fmt.Errorf("signature could not be verified. got: %s, want: %s", plaintext, message)
	}
	return nil
}

// CreateMessageAttestation attests the message using the given public and private key.
// pubKeyEnc: Base64 Encoded Public Key
// privKeyEnc: Base64 Decoded Private Key
// message: Message to attest
func CreateMessageAttestation(pubKeyEnc string, privKeyEnc string, message string) (string, error) {
	// Create a PgpKey from Encoded Public Key
	pgpKey, err := NewPgpKey(privKeyEnc, pubKeyEnc)
	if err != nil {
		return "", errors.Wrap(err, "creating PGP key")
	}
	// First Create a signer Entitiy from public and private keys.
	signer, err := createEntityFromKeys(pgpKey.PublicKey(), pgpKey.PrivateKey())
	if err != nil {
		return "", errors.Wrap(err, "creating entity keys")
	}

	b := new(bytes.Buffer)
	// Armor Decode it.
	armorWriter, errEncode := armor.Encode(b, openpgp.SignatureType, make(map[string]string))
	if errEncode != nil {
		return "", errors.Wrap(err, "encoding data")
	}
	// Finally Sign the Text.
	w, err := openpgp.Sign(armorWriter, signer, nil, &pgpConfig)
	if err != nil {
		return "", errors.Wrap(err, "opengpg signing")
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return "", errors.Wrap(err, "writing signed data")
	}
	w.Close()
	armorWriter.Close()
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func createEntityFromKeys(pubKey *packet.PublicKey, privKey *packet.PrivateKey) (*openpgp.Entity, error) {
	currentTime := pgpConfig.Now()
	uid := packet.NewUserId("", "", "")
	if uid == nil {
		return nil, errors.New("user id field contained invalid characters")
	}

	e := &openpgp.Entity{
		PrimaryKey: pubKey,
		PrivateKey: privKey,
		Identities: make(map[string]*openpgp.Identity),
	}
	isPrimaryId := true
	e.Identities[uid.Id] = &openpgp.Identity{
		Name:   uid.Id,
		UserId: uid,
		SelfSignature: &packet.Signature{
			CreationTime: currentTime,
			SigType:      packet.SigTypePositiveCert,
			PubKeyAlgo:   packet.PubKeyAlgoRSA,
			Hash:         pgpConfig.Hash(),
			IsPrimaryId:  &isPrimaryId,
			FlagsValid:   true,
			FlagSign:     true,
			FlagCertify:  true,
			IssuerKeyId:  &e.PrimaryKey.KeyId,
		},
	}
	err := e.Identities[uid.Id].SelfSignature.SignUserId(uid.Id, e.PrimaryKey, e.PrivateKey, &pgpConfig)
	if err != nil {
		return nil, err
	}

	// Set Config Hash from Config
	hashID, ok := s2k.HashToHashId(pgpConfig.DefaultHash)
	if !ok {
		return nil, fmt.Errorf("tried to convert unknown hash %d", pgpConfig.DefaultHash)
	}
	e.Identities[uid.Id].SelfSignature.PreferredHash = []uint8{hashID}

	// Set Config Cipher from Config
	e.Identities[uid.Id].SelfSignature.PreferredSymmetric = []uint8{uint8(pgpConfig.DefaultCipher)}

	return e, nil
}
