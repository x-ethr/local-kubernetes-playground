package token

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v5"
)

var key *ecdsa.PrivateKey
var pkey *ecdsa.PublicKey

// decode ...
// private argument must be a private pem-encoded key
// public argument must be a public pem-encoded key
func decode(private, public []byte) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	block, _ := pem.Decode(private)
	x509encoded := block.Bytes
	pkey, _ := x509.ParseECPrivateKey(x509encoded)

	publicblock, _ := pem.Decode(public)
	x509encodedpublic := publicblock.Bytes
	genericpublickey, _ := x509.ParsePKIXPublicKey(x509encodedpublic)
	pubkey := genericpublickey.(*ecdsa.PublicKey)

	return pkey, pubkey
}

func init() {
	// secret, e := os.ReadFile("/etc/secrets/jwt-ecdsa-pem")
	// if e != nil {
	// 	slog.Error("Unable to Read Secret from Volume Mount", slog.String("path", "/etc/secrets/jwt-ecdsa-pem"), slog.String("error", e.Error()))
	//
	// 	panic(e)
	// }

	var configurations = make(map[string][]byte)

	e := filepath.WalkDir("/etc/secrets/jwt-ecdsa-pem", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !(d.IsDir()) && d.Type().IsRegular() {
			value, e := os.ReadFile(path)
			if e != nil {
				return e
			}

			configurations[d.Name()] = value
		}

		return nil
	})

	if e != nil {
		slog.Error("Error Walking Secrets Volume Mount", slog.String("error", e.Error()))
		panic(e)
	}

	for key, bytes := range configurations {
		slog.Info("Secret", slog.String("key", key), slog.String("value", string(bytes)))
	}

	pemprivate, pempublic := decode(configurations["ecdsa.private.pem"], configurations["ecdsa.public.pem"])

	key = pemprivate
	pkey = pempublic
}

func Verify(ctx context.Context, t string) (*jwt.Token, error) {
	token, e := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		v, ok := token.Method.(*jwt.SigningMethodECDSA)
		if !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}

		_ = v

		return pkey, nil
	})

	if e != nil {
		slog.WarnContext(ctx, "Error Parsing JWT Token", slog.String("error", e.Error()), slog.String("jwt", t))
		return nil, e
	}

	switch {
	case token.Valid:
		slog.DebugContext(ctx, "Verified Valid Token")
		return token, nil
	case errors.Is(e, jwt.ErrTokenMalformed):
		slog.WarnContext(ctx, "Unable to Verify Malformed String as JWT Token", slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		slog.WarnContext(ctx, "Invalid JWT Signature", slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenExpired):
		slog.WarnContext(ctx, "Expired JWT Token", slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenNotValidYet):
		slog.WarnContext(ctx, "Received a Future, Valid JWT Token", slog.String("error", e.Error()))
	default:
		slog.ErrorContext(ctx, "Unknown Error While Attempting to Validate JWT Token", slog.String("error", e.Error()))
	}

	return nil, e
}
