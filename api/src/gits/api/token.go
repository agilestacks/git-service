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
	cypherBlockLen = aes.BlockSize
	macLen         = sha1.Size
)

var (
	// 2 hex encoding chars per byte
	// 1 block for iv + 2 blocks of data + mac
	deploymentKeyMinHexLen = 2 * ((1+2)*cypherBlockLen + macLen) // 136
	// + 1 extra block for subject
	deploymentKeyMaxHexLen     = deploymentKeyMinHexLen + 2*cypherBlockLen
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

func decodeDeploymentKey(deploymentKeyHex string) (string, string, error) {
	userId := "user-id-that-wont-match-anything"
	subject := ""

	hexLen := len(deploymentKeyHex)
	if hexLen < deploymentKeyMinHexLen || hexLen > deploymentKeyMaxHexLen ||
		(hexLen-2*macLen)%(2*cypherBlockLen) != 0 {

		return userId, subject, fmt.Errorf("Bad deployment key length %d", hexLen)
	}

	deploymentKey, err := hex.DecodeString(deploymentKeyHex)
	if err != nil {
		return userId, subject, err
	}

	if len(deploymentKeySecret) == 0 {
		return userId, subject, fmt.Errorf("Deployment key decoding not initialized")
	}

	mac := deploymentKey[0:macLen]
	encryptedMaterial := deploymentKey[macLen:]

	h := hmac.New(deploymentKeyMacAlg, deploymentKeySecret)
	h.Write(encryptedMaterial)
	expectedMac := h.Sum(nil)
	macErr := fmt.Errorf("Bad MAC: expected %v; got %v", expectedMac, mac)
	if hmac.Equal(expectedMac, mac) {
		macErr = nil
	}

	block, err := aes.NewCipher(deploymentKeyEncryptionKey)
	if err != nil {
		return userId, subject, err
	}
	decrypter := cipher.NewCBCDecrypter(block, deploymentKeyIV)
	paddedMaterial := make([]byte, len(encryptedMaterial))
	decrypter.CryptBlocks(paddedMaterial, encryptedMaterial)

	i := bytes.Index(paddedMaterial, deploymentKeySep)
	if i > 0 && macErr == nil {
		userId = string(paddedMaterial[0:i])
		if i < len(paddedMaterial) {
			rest := paddedMaterial[i:]
			i := bytes.Index(rest, deploymentKeySep)
			if i > 0 {
				subject = string(rest[0:i])
			}
		}
	}

	return userId, subject, macErr
}
