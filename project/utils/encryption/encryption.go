package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ingoxx/go-gin/project/config"
	ginErr "github.com/ingoxx/go-gin/project/errors"
	"golang.org/x/crypto/bcrypt"
	"os"
)

type DataEncryption struct {
	password string
	user     string
}

func (e *DataEncryption) EncryptionPwd() (string, error) {
	password := e.user + e.password + config.Sign
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil || len(hashedPassword) == 0 {
		return "", ginErr.NewEncryptionDataError(err.Error())
	}

	return string(hashedPassword), nil
}

func (e *DataEncryption) DecryptionPwd(hashPwd string) (err error) {
	password := e.user + e.password + config.Sign
	err = bcrypt.CompareHashAndPassword([]byte(hashPwd), []byte(password))
	if err != nil {
		return
	}

	return
}

func NewDataEncryption(u, p string) *DataEncryption {
	return &DataEncryption{
		user:     u,
		password: p,
	}
}

type KeyPwdEncryption struct {
	data     string
	dataType uint32
}

func (eke *KeyPwdEncryption) getPwdString() (kc string, err error) {
	if eke.dataType == 1 {
		return eke.data, nil
	} else if eke.dataType == 2 {
		data, err := os.ReadFile(eke.data)
		if err != nil {
			return "", err
		}
		return string(data), nil
	} else {
		return "", errors.New("未知类型, 无法加密")
	}
}

func (eke *KeyPwdEncryption) getKey() []byte {
	key := []byte(config.Key)
	return key[:32]
}

func (eke *KeyPwdEncryption) Encryption() (ks string, err error) {
	key := eke.getKey()
	enData, err := eke.getPwdString()
	if err != nil {
		return
	}

	// 创建 AES 密码块
	block, err := aes.NewCipher(eke.getKey())
	if err != nil {
		return "", fmt.Errorf("创建 AES 失败: %w", err)
	}

	// 生成 IV（初始化向量），长度必须与 block.BlockSize() 相同
	iv := key[:block.BlockSize()]

	plainBytes := []byte(enData)
	ciphertext := make([]byte, len(plainBytes))

	// 使用 CTR 模式加密
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, plainBytes)

	// 返回 Base64 编码的加密结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (eke *KeyPwdEncryption) Decryption() (ks string, err error) {
	key := eke.getKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建 AES 失败: %w", err)
	}

	// 生成 IV（初始化向量），长度必须与 block.BlockSize() 一致
	iv := key[:block.BlockSize()]

	// 解码 Base64
	cipherText, err := base64.StdEncoding.DecodeString(eke.data)
	if err != nil {
		return "", fmt.Errorf("Base64 解码失败: %w", err)
	}

	plainText := make([]byte, len(cipherText))

	// 创建 CTR 解密流（CTR 其实是对称的，只需再次 XOR 即可解密）
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plainText, cipherText)

	return string(plainText), nil
}

func NewKeyPwdEncryption(data string, dataType uint32) *KeyPwdEncryption {
	return &KeyPwdEncryption{
		data:     data,
		dataType: dataType,
	}
}
