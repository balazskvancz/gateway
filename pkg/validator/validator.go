package validator

import (
	"fmt"
	"net/http"

	"github.com/balazskvancz/gateway/pkg/gcontext"
	"github.com/balazskvancz/gateway/pkg/utils"
)

type Validator struct {
	FieldName string

	SecretKey string
	HashedKey string
}

func New(field, key string) *Validator {
	if field == "" || key == "" {
		return nil
	}

	hash := utils.CreateHash(key)

	return &Validator{
		FieldName: field,
		SecretKey: key,
		HashedKey: hash,
	}
}

// Basic validation for signature in header.
func (v *Validator) ValidateHeader(ctx *gcontext.GContext) bool {
	headerSignature := ctx.GetRequestHeader(v.FieldName)

	if headerSignature == "" {
		return false
	}

	decoded := utils.DecodeB64(headerSignature)

	if decoded == "" {
		return false
	}

	// In case of @GET, we check for only key hash equality.
	if ctx.GetRequestMethod() == http.MethodGet {
		return decoded == v.HashedKey
	}

	b, err := ctx.GetRawBody()

	if err != nil {
		fmt.Printf("[VALIDATOR]: reading context body error: %v\n", err)

		return false
	}

	// The json serialized string and the key concatenated.
	computedOriginal := string(b) + v.SecretKey
	computedHash := utils.CreateHash(computedOriginal)

	return decoded == computedHash
}
