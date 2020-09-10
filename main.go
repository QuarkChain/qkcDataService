package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/QuarkChain/qkcDataService/controllers"
	_ "github.com/QuarkChain/qkcDataService/routers"
	"github.com/astaxie/beego"
	"io"
)

var (
	opType  = flag.String("type", "", "encrypt or run")
	private = flag.String("private_key", "", "private key")
	pass    = flag.String("password", "", "password")
	host    = flag.String("host", "", "host")
)

func Init() {
	controllers.SDK = controllers.NewQKCSDK(decrypt(), *host)
}

func check() {
	if *pass == "" {
		panic("password should not empty")
	}
	if *private == "" {
		panic("private should not empty")
	}
}

func transPass(pass string) []byte {
	h := sha256.New()
	h.Write([]byte(pass))
	return h.Sum(nil)
}
func encrypt() {
	check()
	r := AesEncryptCFB([]byte(*private), transPass(*pass))
	fmt.Println("加密后的密钥为", hex.EncodeToString(r))
}

func decrypt() string {
	if *pass == "" {
		return *private
	}
	pByte, err := hex.DecodeString(*private)
	if err != nil {
		panic(encrypt)
	}
	return string(AesDecryptCFB(pByte, transPass(*pass)))
}

// =================== CFB ======================
func AesEncryptCFB(origData []byte, key []byte) (encrypted []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	encrypted = make([]byte, aes.BlockSize+len(origData))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], origData)
	return encrypted
}
func AesDecryptCFB(encrypted []byte, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)
	if len(encrypted) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)
	return encrypted
}
func main() {
	flag.Parse()
	switch *opType {
	case "":
		Init()
		beego.Run()
	case "encrypt":
		encrypt()
	}
}
