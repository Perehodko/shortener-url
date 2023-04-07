package workwithcookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var secret = "secretForEncrypt"

// EncryptedUUID вычисление HMAC-SHA256 для переданной строки
func EncryptedUUID(data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	EncryptedUUID := hex.EncodeToString(h.Sum(nil))
	return EncryptedUUID
}

// CheckTokenIsValid проверка хеша
func CheckTokenIsValid(data string, hash string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	sign, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}
	isEqual := hmac.Equal(sign, h.Sum(nil))
	return isEqual
}

// ExtractUID извлекает из куки uid пользователя и валидирует его.
// Если хэш корректный, возвращает uid.
// Если куки нет или хеш невалидный - генерирует новый uid
func ExtractUID(cookies []*http.Cookie) (string, error) {
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			parts := strings.Split(cookie.Value, ":")
			if len(parts) != 2 {
				return "", errors.New("no cookie with need name")
			}
			UUID, hash := parts[0], parts[1]
			if CheckTokenIsValid(UUID, hash) {
				return UUID, nil
			}
			return "", errors.New("invalid cookie")
		}
	}
	return "", errors.New("no cookie")
}

// SetUUIDCookie сохраняет в куку uuid пользователя вместе с его hmac
func SetUUIDCookie(w http.ResponseWriter, uid string) {
	UUIDEncrypted := fmt.Sprintf("%s:%s", uid, EncryptedUUID(uid))

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  UUIDEncrypted,
		MaxAge: 10000,
	})
}
