package prescription

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"

	"github.com/gowebpki/jcs"
)

const (
	algorithmJCSSha512 = "JCSSHA512" // SHA512 of Canonicalized JSON Document (RFC8785)
)

type IntegrityHash struct {
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
}

func NewIntegrityAttributesFromRevisionCreate(create RevisionCreate) DataAttributes {
	return create.DataAttributes
}

func NewIntegrityAttributesFromRevision(revision Revision) DataAttributes {
	return revision.Attributes.DataAttributes
}

// generateJCSSha512 computes the hex encoded sha512 hash of the canonicalized json of prescription attributes
func generateJCSSha512(attributes DataAttributes) (string, error) {
	// marshal the relevant attributes to json
	payload, err := json.Marshal(attributes)
	if err != nil {
		return "", err
	}

	// canonicalize the json document
	payload, err = jcs.Transform(payload)
	if err != nil {
		return "", err
	}

	// compute the sha512 hash
	hasher := sha512.New()
	hasher.Write(payload)
	hash := hasher.Sum(nil)

	return fmt.Sprintf("%x", hash), nil
}

func MustGenerateIntegrityHash(attributes DataAttributes) IntegrityHash {
	hash, err := generateJCSSha512(attributes)
	if err != nil {
		panic(err)
	}

	return IntegrityHash{
		Hash:      hash,
		Algorithm: algorithmJCSSha512,
	}
}

func VerifyRevisionIntegrityHash(revision Revision) error {
	attrs := NewIntegrityAttributesFromRevision(revision)
	hash := MustGenerateIntegrityHash(attrs)
	if revision.IntegrityHash == nil {
		return fmt.Errorf("revision integrity hash missing")
	}
	if !hashesMatch(*revision.IntegrityHash, hash) {
		return fmt.Errorf("revision attributes do not match integrity hash")
	}
	return nil
}

func hashesMatch(a, b IntegrityHash) bool {
	return a.Algorithm == b.Algorithm && a.Hash == b.Hash
}
