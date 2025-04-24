package assertion

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/bas-d/appattest/authenticator"
	"github.com/ugorji/go/codec"

	"github.com/bas-d/appattest/utils"
)

type ClientData struct {
	Challenge string `json:"challenge"`
}

type AuthenticatorAssertionResponse struct {
	ClientDataJSON ClientData
	RawClientData  utils.URLEncodedBase64 `json:"clientData"`
	Assertion      utils.URLEncodedBase64 `json:"assertion"`
}

type Assertion struct {
	AuthenticatorData    authenticator.AuthenticatorData
	RawAuthenticatorData []byte `json:"authenticatorData"`
	Signature            []byte `json:"signature"`
}

func (aar *AuthenticatorAssertionResponse) Verify(storedChallenge string, relyingPartyID string, previousCounter uint32, publicKey []byte) (uint32, error) {
	a, err := aar.parse()
	if err != nil {
		return 0, err
	}

	// 1. Compute clientDataHash as the SHA256 hash of clientData.
	clientDataHash := sha256.Sum256(aar.RawClientData)

	// 2. Concatenate authenticatorData and clientDataHash and apply a SHA256 hash over the result to form nonce.
	nonceData := append(a.RawAuthenticatorData, clientDataHash[:]...)
	nonce := sha256.Sum256(nonceData)

	// 3. Use the public key that you stored from the attestation object to verify that the assertion’s signature is valid for nonce.
	x, y := elliptic.Unmarshal(elliptic.P256(), publicKey)
	if x == nil {
		return 0, utils.ErrParsingData.WithDetails("Failed to parse the public key")
	}
	pubkey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	nonceHash := sha256.Sum256(nonce[:])
	valid := ecdsa.VerifyASN1(pubkey, nonceHash[:], a.Signature)
	if !valid {
		return 0, utils.ErrAssertionSignature.WithDetails("Error validating the assertion signature.\n")
	}

	// 4. Compute the SHA256 hash of the client’s App ID, and verify that it matches the RP ID in the authenticator data.
	rpIDHash := sha256.Sum256([]byte(relyingPartyID))
	if !bytes.Equal(a.AuthenticatorData.RPIDHash[:], rpIDHash[:]) {
		return 0, utils.ErrVerification.WithDetails(fmt.Sprintf("RP Hash mismatch. Expected %x and Received %x\n", a.AuthenticatorData.RPIDHash, rpIDHash))
	}

	// 5. Verify that the authenticator data’s counter value is greater than the value from the previous assertion, or greater than 0 on the first assertion.
	if a.AuthenticatorData.Counter <= previousCounter {
		return 0, utils.ErrVerification.WithDetails(fmt.Sprintf("Counter was not not greater than previous  %d\n", a.AuthenticatorData.Counter))
	}

	// 6. Verify that the challenge embedded in the client data matches the earlier challenge to the client.
	if storedChallenge != aar.ClientDataJSON.Challenge {
		err := utils.ErrChallengeMismatch.WithDetails("Error validating challenge")
		return 0, err.WithDetails(fmt.Sprintf("Expected b Value: %#v\nReceived b: %#v\n", storedChallenge, aar.ClientDataJSON))
	}

	return a.AuthenticatorData.Counter, nil
}

func (aar *AuthenticatorAssertionResponse) parse() (*Assertion, error) {
	var a Assertion

	cborHandler := codec.CborHandle{}

	// Decode the attestation data with unmarshalled auth data
	err := codec.NewDecoderBytes(aar.Assertion, &cborHandler).Decode(&a)
	if err != nil {
		return nil, utils.ErrParsingData.WithDetails(err.Error())
	}

	err = a.AuthenticatorData.Unmarshal(a.RawAuthenticatorData)
	if err != nil {
		return nil, fmt.Errorf("error decoding auth data: %v", err)
	}

	if err := json.Unmarshal(aar.RawClientData, &aar.ClientDataJSON); err != nil {
		return nil, fmt.Errorf("error decoding client data: %v", err)
	}

	return &a, nil
}
