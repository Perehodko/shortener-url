package work_with_cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
)

func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

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

const uidCookieName = "SHORTENER_UID"

var (
	ErrInvalidCookieValue  = errors.New("invalid cookie value")
	ErrInvalidCookieDigest = errors.New("invalid cookie digest")
	ErrNoCookie            = errors.New("no cookie")
)

var secret = "secretForEncrypt" // Прочитать из env/конфига

// EncryptedUUID вычисление HMAC-SHA256 для переданной строки
func EncryptedUUID(data string) string {
	secret, _ := GenerateRandom(32)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// CheckTokenIsValid проверка хеша
func CheckTokenIsValid(data string, hash string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	sign, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}
	return hmac.Equal(sign, h.Sum(nil))
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
		}
	}
	return "", errors.New("no cookie")
}

// SetUUIDCookie сохраняет в куку uuid пользователя вместе с его hmac
func SetUUIDCookie(w http.ResponseWriter, uid string) {
	UUIDEncrypted := fmt.Sprintf("%s:%s", uid, EncryptedUUID(uid))
	fmt.Println("UUIDEncrypted", EncryptedUUID(uid))

	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: UUIDEncrypted,
	})
}
