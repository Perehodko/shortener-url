package workwithcookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

//// EncryptedUUID шифрует UUID и возвращает зашифрованный массив байт
//func EncryptedUUID() ([]byte, error, []byte, string, []byte) {
//	UUID := uuid.New()
//	//fmt.Println(UUID.String(), UUID)
//
//	src := []byte(UUID.String()) // данные, которые хотим зашифровать
//	fmt.Printf("original: %s\n", src)
//
//	// будем использовать AES256, создав ключ длиной 32 байта
//	key, err := GenerateRandom(2 * aes.BlockSize) // ключ шифрования
//	if err != nil {
//		fmt.Printf("error: %v\n", err)
//	}
//
//	aesblock, err := aes.NewCipher(key)
//	if err != nil {
//		fmt.Printf("error: %v\n", err)
//	}
//
//	aesgcm, err := cipher.NewGCM(aesblock)
//	if err != nil {
//		fmt.Printf("error: %v\n", err)
//	}
//
//	// создаём вектор инициализации
//	nonce, err := GenerateRandom(aesgcm.NonceSize())
//	if err != nil {
//		fmt.Printf("error: %v\n", err)
//	}
//	fmt.Println("nonce", nonce)
//
//	encryptedUUID := aesgcm.Seal(nil, nonce, src, nil) // зашифровываем
//	//fmt.Printf("encrypted: %x\n", encryptedUUID)
//
//	//encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)
//	//fmt.Println(encryptedUUIDStr)
//
//	return encryptedUUID, nil, key, UUID.String(), nonce
//}
//
//// CheckCookieExist проверяет есть ли уже куки и возвращает их
//func CheckCookieExist(r *http.Request) string {
//	cookie, err := r.Cookie("session")
//	var sessionCookie string
//	if err == nil {
//		sessionCookie = cookie.Value
//	} else if err != http.ErrNoCookie {
//		log.Println(err)
//	}
//	return sessionCookie
//}
//
//// CheckKeyIsValid проверяет валиднсть куки (расшифровывает ее и сравнивает с оригиналом)
//func CheckKeyIsValid(key []byte, encryptedUUID []byte, UUID string, nonce []byte) bool {
//	// получаем cipher.Block
//	aesblock, err := aes.NewCipher(key)
//	if err != nil {
//		fmt.Printf("error: %v\n", err)
//	}
//	aesgcm, err := cipher.NewGCM(aesblock)
//	if err != nil {
//		panic(err)
//	}
//
//	src2, err := aesgcm.Open(nil, nonce, encryptedUUID, nil) // расшифровываем
//	if err != nil {
//		fmt.Printf("error checkKeyIsValid: %v\n", err)
//	}
//
//	if UUID == string(src2) {
//		return true
//	} else {
//		return false
//	}
//}

//
//// ExtractUID извлекает из куки uid пользователя и валидирует его.
//// Если хэш корректный, возвращает uid.
//// Если куки нет или хеш невалидный - генерирует новый uid
//func ExtractUID(cookies []*http.Cookie) (string, error) {
//	_, err, key, UUID, nonce := EncryptedUUID()
//	if err != nil {
//		return "", err
//	}
//
//	for _, cookie := range cookies {
//		if cookie.Name == "session" {
//			parts := strings.Split(cookie.Value, ":")
//			if len(parts) != 2 {
//				return "", errors.New("invalid cookie value")
//			}
//			encryptedUID := parts[1]
//			//if checkHash(uid, hash) {
//			//	return uid, nil
//			//}
//			isValid := CheckKeyIsValid(key, []byte(encryptedUID), UUID, nonce)
//			if isValid {
//				return encryptedUID, nil
//			}
//			return "", errors.New("invalid cookie digest")
//		}
//	}
//}

//// SetUUIDCookie сохраняет в куку uid пользователя вместе с его hmac
//func SetUUIDCookie(w http.ResponseWriter, uid string) {
//	encryptedUUID, _, _, _, _ := EncryptedUUID()
//	encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)
//
//	http.SetCookie(w, &http.Cookie{
//		Name:  "session",
//		Value: encryptedUUIDStr,
//	})
//}

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
				return "", errors.New("no cookie with session name")
			}
			UUID, hash := parts[0], parts[1]
			if CheckTokenIsValid(UUID, hash) {
				return UUID, nil
			}
			return "", errors.New("invalid cookie digest")
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

const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

// String генерирует случайную строку заданной длинны,
// содержащую букво-циферную последовательность символов.
//func String(length int) string {
//	if length < 0 {
//		return ""
//	}
//	b := make([]byte, length)
//	for i := range b {
//		b[i] = charSet[seededRand.Intn(len(charSet))]
//	}
//	return string(b)
//}

// UserID генерирует uid пользователя.
// В будущем лучше заменить на https://pkg.go.dev/github.com/google/uuid
func UserID() string {
	return GenerateRandomString(32)
}

func GenerateRandomString(size int) string {
	if size < 0 {
		return ""
	}
	b := make([]byte, size)
	for i := range b {
		b[i] = charSet[seededRand.Intn(len(charSet))]
	}
	return string(b)

}

func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
