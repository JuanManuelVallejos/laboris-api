// Package webhook verifica la firma de los webhooks entrantes de Clerk, que
// usan el esquema estándar de Svix. clerk-sdk-go v2 no trae un verificador de
// firma (su paquete svixwebhook solo administra la config de Clerk, no valida
// payloads entrantes), así que se reimplementa a mano en vez de sumar la
// dependencia github.com/svix/svix-webhooks para esto único.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const timestampTolerance = 5 * time.Minute

// Verify reproduce el esquema de firma de Svix: HMAC-SHA256 sobre
// "{id}.{timestamp}.{body}" con el secreto whsec_..., comparado contra cada
// firma "v1,<base64>" que traiga el header svix-signature (puede traer más
// de una, separadas por espacio, durante una rotación de secreto).
func Verify(secret, id, timestamp, signatureHeader string, body []byte) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("svix-timestamp inválido")
	}
	if age := time.Since(time.Unix(ts, 0)); age > timestampTolerance || age < -timestampTolerance {
		return errors.New("svix-timestamp fuera de tolerancia")
	}

	secretBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "whsec_"))
	if err != nil {
		return errors.New("secreto de webhook inválido")
	}

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(fmt.Sprintf("%s.%s.%s", id, timestamp, body)))
	expected := mac.Sum(nil)

	for _, part := range strings.Split(signatureHeader, " ") {
		v, sig, found := strings.Cut(part, ",")
		if !found || v != "v1" {
			continue
		}
		got, err := base64.StdEncoding.DecodeString(sig)
		if err != nil {
			continue
		}
		if hmac.Equal(got, expected) {
			return nil
		}
	}
	return errors.New("firma inválida")
}
