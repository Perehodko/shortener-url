package work_with_cookie

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func CheckCookieExist(r *http.Request) string {
	cookie, err := r.Cookie("session")
	var sessionCookie string
	if err == nil {
		sessionCookie = cookie.Value
	} else if err != http.ErrNoCookie {
		log.Println(err)
	}
	return sessionCookie
}

func CheckKeyIsValid(key []byte, encryptedUUID []byte, UUID string, nonce []byte) bool {
	//fmt.Println("checkKeyIsValid - encryptedUUID", encryptedUUID)
	//receive := fmt.Sprintf("%s", encryptedUUID)
	//fmt.Println("checkKeyIsValid - encryptedUUID", receive)

	// получаем cipher.Block
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		panic(err)
	}

	//dst := aesgcm.Seal(nil, nonce, []byte(UUID), nil) // зашифровываем
	//fmt.Printf("encrypted: %x\n", dst)
	//
	//fmt.Println("receive, dst", encryptedUUID, dst)
	//dst_2 := fmt.Sprintf("%s", dst)
	//
	//fmt.Println("receive, dst", receive, dst_2)

	src2, err := aesgcm.Open(nil, nonce, encryptedUUID, nil) // расшифровываем
	if err != nil {
		fmt.Printf("error checkKeyIsValid: %v\n", err)
	}
	fmt.Println("расшифррованный UUID ", src2)
	encryptedUUIDStr := fmt.Sprintf("%s", src2)
	fmt.Println("расшифррованный UUID2 ", encryptedUUIDStr)
	fmt.Println("UUID == string(src2)???", UUID == string(src2))

	if UUID == string(src2) {
		return true
	} else {
		return false
	}
}
