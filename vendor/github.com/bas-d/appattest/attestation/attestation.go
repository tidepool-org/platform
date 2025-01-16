package attestation

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"fmt"

	"github.com/bas-d/appattest/authenticator"
	"github.com/bas-d/appattest/utils"
	"github.com/ugorji/go/codec"
)

const APPLE_ROOT_CERT = `-----BEGIN CERTIFICATE-----
MIICITCCAaegAwIBAgIQC/O+DvHN0uD7jG5yH2IXmDAKBggqhkjOPQQDAzBSMSYw
JAYDVQQDDB1BcHBsZSBBcHAgQXR0ZXN0YXRpb24gUm9vdCBDQTETMBEGA1UECgwK
QXBwbGUgSW5jLjETMBEGA1UECAwKQ2FsaWZvcm5pYTAeFw0yMDAzMTgxODMyNTNa
Fw00NTAzMTUwMDAwMDBaMFIxJjAkBgNVBAMMHUFwcGxlIEFwcCBBdHRlc3RhdGlv
biBSb290IENBMRMwEQYDVQQKDApBcHBsZSBJbmMuMRMwEQYDVQQIDApDYWxpZm9y
bmlhMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAERTHhmLW07ATaFQIEVwTtT4dyctdh
NbJhFs/Ii2FdCgAHGbpphY3+d8qjuDngIN3WVhQUBHAoMeQ/cLiP1sOUtgjqK9au
Yen1mMEvRq9Sk3Jm5X8U62H+xTD3FE9TgS41o0IwQDAPBgNVHRMBAf8EBTADAQH/
MB0GA1UdDgQWBBSskRBTM72+aEH/pwyp5frq5eWKoTAOBgNVHQ8BAf8EBAMCAQYw
CgYIKoZIzj0EAwMDaAAwZQIwQgFGnByvsiVbpTKwSga0kP0e8EeDS4+sQmTvb7vn
53O5+FRXgeLhpJ06ysC5PrOyAjEAp5U4xDgEgllF7En3VcE3iexZZtKeYnpqtijV
oyFraWVIyd/dganmrduC1bmTBGwD
-----END CERTIFICATE-----`

const attestationKey = "apple-appattest"

type AuthenticatorAttestationResponse struct {
	ClientData        utils.URLEncodedBase64 `json:"clientData"`
	KeyID             string                 `json:"keyID"`
	AttestationObject utils.URLEncodedBase64 `json:"attestationObject"`
}

type AttestationObject struct {
	AuthData     authenticator.AuthenticatorData
	RawAuthData  []byte                 `json:"authData"`
	Format       string                 `json:"fmt"`
	AttStatement map[string]interface{} `json:"attStmt,omitempty"`
}

func (aar *AuthenticatorAttestationResponse) Verify(appID string, production bool) ([]byte, []byte, error) {
	a, err := aar.parse()
	if err != nil {
		return nil, nil, err
	}

	// Compute clientDataHash as the SHA256 hash of clientData.
	clientDataHash := sha256.Sum256(aar.ClientData)

	// Check if we have the right format.
	if a.Format != attestationKey {
		return nil, nil, utils.ErrAttestationFormat.WithDetails(fmt.Sprintf("Wrong attestation format unsupported: %s", a.Format))
	}

	// Decode the key ID
	keyIdData, err := base64.StdEncoding.DecodeString(aar.KeyID)
	if err != nil {
		return nil, nil, utils.ErrParsingData.WithDetails(fmt.Sprintf("The KeyID was not valid base64: %s", aar.KeyID))
	}

	// Handle Steps 6 through 9
	// 6. Compute the SHA256 hash of your app’s App ID
	appIDHash := sha256.Sum256([]byte(appID))
	authDataVerificationError := a.AuthData.Verify(appIDHash[:], keyIdData, production)
	if authDataVerificationError != nil {
		return nil, nil, authDataVerificationError
	}

	// Handle step 1 through 5
	return verifyAttestation(*a, clientDataHash[:], keyIdData)
}

func (aar *AuthenticatorAttestationResponse) parse() (*AttestationObject, error) {
	var a AttestationObject

	cborHandler := codec.CborHandle{}

	err := codec.NewDecoderBytes(aar.AttestationObject, &cborHandler).Decode(&a)
	if err != nil {
		return nil, utils.ErrParsingData.WithDetails(err.Error())
	}

	err = a.AuthData.Unmarshal(a.RawAuthData)
	if err != nil {
		return nil, fmt.Errorf("error decoding auth data: %v", err)
	}

	if !a.AuthData.Flags.HasAttestedCredentialData() {
		return nil, utils.ErrAttestationFormat.WithDetails("Attestation missing attested credential data flag")
	}

	return &a, nil
}

func verifyAttestation(att AttestationObject, clientDataHash, keyID []byte) ([]byte, []byte, error) {
	// Validate according to https://developer.apple.com/documentation/devicecheck/validating_apps_that_connect_to_your_server
	// Create certificate pool with the Apple Root cert.

	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()

	// Add Apple root Cert
	ok := roots.AppendCertsFromPEM([]byte(APPLE_ROOT_CERT))
	if !ok {
		return nil, nil, utils.ErrAttestationFormat.WithDetails("Error adding root certificate to pool.")
	}

	x5c, x509present := att.AttStatement["x5c"].([]interface{})
	if !x509present {
		return nil, nil, utils.ErrAttestationFormat.WithDetails("Error retrieving x5c value")
	}

	receipt, receiptPresent := att.AttStatement["receipt"].([]byte)
	if !receiptPresent {
		return nil, nil, utils.ErrAttestationFormat.WithDetails("Error retreiving receipt value")
	}

	for _, c := range x5c {
		cb, cv := c.([]byte)
		if !cv {
			return nil, nil, utils.ErrAttestationCertificate.WithDetails("Error getting certificate from x5c cert chain 1")
		}
		ct, err := x509.ParseCertificate(cb)
		if err != nil {
			return nil, nil, utils.ErrAttestationCertificate.WithDetails(fmt.Sprintf("Error parsing certificate from ASN.1 data: %+v", err))
		}
		if ct.IsCA {
			intermediates.AddCert(ct)
		}
	}

	credCertBytes, valid := x5c[0].([]byte)
	if !valid {
		return nil, nil, utils.ErrAttestationCertificate.WithDetails("Error getting certificate from x5c cert chain 2")
	}

	credCert, err := x509.ParseCertificate(credCertBytes)
	if err != nil {
		return nil, nil, utils.ErrAttestationCertificate.WithDetails(fmt.Sprintf("Error parsing certificate from ASN.1 data: %+v", err))
	}

	// Create verification options.
	verifyOptions := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	}

	// 1. Verify that the x5c array contains the intermediate and leaf certificates for App Attest,
	// starting from the credential certificate stored in the first data buffer in the array (credcert).
	// Verify the validity of the certificates using Apple’s root certificate.
	_, err = credCert.Verify(verifyOptions)
	if err != nil {
		return nil, nil, utils.ErrAttestationCertificate.WithDetails(fmt.Sprintf("Invalid certificate %+v", err))
	}

	// 2. Create clientDataHash as the SHA256 hash of the one-time challenge sent to your app before performing the attestation,
	// and append that hash to the end of the authenticator data (authData from the decoded object).
	nonceData := append(att.RawAuthData, clientDataHash...)

	// 3. Generate a new SHA256 hash of the composite item to create nonce.
	nonce := sha256.Sum256(nonceData)

	// 4. Obtain the value of the credCert extension with OID 1.2.840.113635.100.8.2, which is a DER-encoded ASN.1 sequence.
	// Decode the sequence and extract the single octet string that it contains.
	// Verify that the string equals nonce.
	credCertOID := asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 2}
	var credCertId []byte
	for _, extension := range credCert.Extensions {
		if extension.Id.Equal(credCertOID) {
			credCertId = extension.Value
		}
	}

	if len(credCertId) <= 0 {
		return nil, nil, utils.ErrInvalidAttestation.WithDetails("Certificate did not contain credCert extension")
	}
	var unMarshalledCredCertOctet []asn1.RawValue
	var unMarshalledCredCert asn1.RawValue
	asn1.Unmarshal(credCertId, &unMarshalledCredCertOctet)
	asn1.Unmarshal(unMarshalledCredCertOctet[0].Bytes, &unMarshalledCredCert)
	if !bytes.Equal(nonce[:], unMarshalledCredCert.Bytes) {
		return nil, nil, utils.ErrInvalidAttestation.WithDetails("Certificate CredCert extension does not match nonce.")
	}

	// 5. Create the SHA256 hash of the public key in credCert, and verify that it matches the key identifier from your app.
	var publicKeyBytes []byte
	switch pub := credCert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		publicKeyBytes = elliptic.Marshal(pub.Curve, pub.X, pub.Y)
		pubKeyHash := sha256.Sum256(publicKeyBytes)
		if !bytes.Equal(pubKeyHash[:], keyID) {
			return nil, nil, utils.ErrInvalidAttestation.WithDetails("The key id is not a valid SHA256 hash of the certificate public key.")
		}
	default:
		return nil, nil, utils.ErrInvalidAttestation.WithDetails("Wrong algorithm")
	}

	// Return x963-encoded public key and receipt.
	return publicKeyBytes, receipt, nil
}
