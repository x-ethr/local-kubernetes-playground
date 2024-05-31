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
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var h []byte
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

	e = filepath.WalkDir("/etc/secrets/jwt-signing-token", func(path string, d fs.DirEntry, err error) error {
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

	h = configurations["jwt-signing-token"]

	key = pemprivate
	pkey = pempublic
}

func Create(ctx context.Context, token *jwt.Token) (string, error) {
	hmactoken := jwt.New(jwt.SigningMethodHS256)
	hmacclaims := hmactoken.Claims.(jwt.MapClaims)
	for key, element := range token.Claims.(jwt.MapClaims) {
		hmacclaims[key] = element
	}

	expiration := time.Now().Add(time.Hour * 1).Unix()
	hmacclaims["exp"] = expiration

	hmactokenstring, e := hmactoken.SignedString(h)
	if e != nil {
		slog.WarnContext(ctx, "Error Signing JWT Token", slog.Any("token", token), slog.String("error", e.Error()))

		return "", e
	}

	return hmactokenstring, nil
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
		slog.WarnContext(ctx, "Error Parsing JWT Token", slog.String("error", e.Error()))
		return nil, e
	}

	switch {
	case token.Valid:
		slog.DebugContext(ctx, "Verified Valid Token", slog.Any("token", token))
		return token, nil
	case errors.Is(e, jwt.ErrTokenMalformed):
		slog.WarnContext(ctx, "Unable to Verify Malformed String as JWT Token", slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		slog.WarnContext(ctx, "Invalid JWT Signature", slog.Any("token", token), slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenExpired):
		slog.WarnContext(ctx, "Expired JWT Token", slog.Any("token", token), slog.String("error", e.Error()))
	case errors.Is(e, jwt.ErrTokenNotValidYet):
		slog.WarnContext(ctx, "Received a Future, Valid JWT Token", slog.Any("token", token), slog.String("error", e.Error()))
	default:
		slog.ErrorContext(ctx, "Unknown Error While Attempting to Validate JWT Token", slog.Any("token", token), slog.String("error", e.Error()))
	}

	return nil, e
}
