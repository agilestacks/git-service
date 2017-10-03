package api

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"

	"golang.org/x/crypto/pbkdf2"

	"gits/config"
)

const (
	deploymentKeyHexLen   = 2 * (16 + 2*16 + 20) // 136
	deploymentKeyMacLen   = 20
	deploymentKeyBlockLen = 16 + 2*16
)

var (
	deploymentKeyMacAlg        = sha1.New
	deploymentKeySalt          = []byte("Git")
	deploymentKeyIV            = []byte("gah4ixaXuuShe4qu")
	deploymentKeySep           = []byte("|")
	deploymentKeySecret        []byte
	deploymentKeyEncryptionKey []byte
)

func Init() {
	if (config.GitApiSecret == "" || config.HubApiSecret == "") && config.Verbose {
		log.Print("Either Git or Hub API secret is not set - user's deployment keys won't work")
		return
	}
	deploymentKeySecret = []byte(fmt.Sprintf("%s|%s", config.HubApiSecret, config.GitApiSecret))
	deploymentKeyEncryptionKey = pbkdf2.Key(deploymentKeySecret, deploymentKeySalt, 4096, 32, deploymentKeyMacAlg)
}

func decodeDeploymentKey(deploymentKeyHex string) (string, error) {
	userId := "user-id-that-wont-match-anything"

	if len(deploymentKeyHex) != deploymentKeyHexLen {
		return userId, fmt.Errorf("Bad deployment key length %d", len(deploymentKeyHex))
	}

	deploymentKey, err := hex.DecodeString(deploymentKeyHex)
	if err != nil {
		return userId, err
	}

	if len(deploymentKeySecret) == 0 {
		return userId, fmt.Errorf("Deployment key decoding not initialized")
	}

	mac := deploymentKey[0:deploymentKeyMacLen]
	encryptedUserId := deploymentKey[deploymentKeyMacLen:]

	h := hmac.New(deploymentKeyMacAlg, deploymentKeySecret)
	h.Write(encryptedUserId)
	expectedMac := h.Sum(nil)
	var macErr error
	if !hmac.Equal(expectedMac, mac) {
		macErr = fmt.Errorf("Bad MAC: expected %v; got %v", expectedMac, mac)
	}

	block, err := aes.NewCipher(deploymentKeyEncryptionKey)
	if err != nil {
		return userId, err
	}
	decrypter := cipher.NewCBCDecrypter(block, deploymentKeyIV)
	paddedUserId := make([]byte, deploymentKeyBlockLen)
	decrypter.CryptBlocks(paddedUserId, encryptedUserId)

	sepIndex := bytes.Index(paddedUserId, deploymentKeySep)
	if sepIndex > 0 {
		userId = string(paddedUserId[0:sepIndex])
	}

	return userId, macErr
}
