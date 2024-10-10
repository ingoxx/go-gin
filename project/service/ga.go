package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//请查看规范文档 https://github.com/google/google-authenticator/wiki/Key-Uri-Format

type GoogleAuthenticator struct {
	Secret string
	Expire uint64
	Digits int
}

func (m *GoogleAuthenticator) GaCode() (code string, err error) {
	count := uint64(time.Now().Unix()) / m.Expire
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(m.Secret)
	if err != nil {

		return
	}
	codeInt := hotp(key, count, m.Digits)
	intFormat := fmt.Sprintf("%%0%dd", m.Digits)
	return fmt.Sprintf(intFormat, codeInt), nil
}

func (m *GoogleAuthenticator) QrUrl(label, user string) (qr string) {
	m.CreateSecret(user)
	flabel := url.QueryEscape(label)

	//otpauth://totp/ACME%20Co:john.doe@email.com?secret=HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ&issuer=ACME%20Co&algorithm=SHA1&digits=6&period=30
	qr = fmt.Sprintf(`otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d`, user, m.Secret, flabel, m.Digits, m.Expire)
	return
}

func (m *GoogleAuthenticator) CreateSecret(u string) {
	date := strconv.Itoa(int(time.Now().Nanosecond()))
	data := "ajksduk912J3KDAKJKASD" + u + date
	hash := sha1.New()
	hash.Write([]byte(data))
	nd := hash.Sum(nil)
	nnd := hex.EncodeToString(nd)
	key := base32.StdEncoding.EncodeToString([]byte(nnd))
	sha1String := strings.Split(key, "=")
	m.Secret = sha1String[0]
}

func NewGoogleAuthenticator(key string) *GoogleAuthenticator {
	return &GoogleAuthenticator{
		Secret: key,
		Expire: 30,
		Digits: 6,
	}
}

func hotp(key []byte, counter uint64, digits int) int {
	//只支持sha1
	h := hmac.New(sha1.New, key)
	binary.Write(h, binary.BigEndian, counter)
	sum := h.Sum(nil)
	v := binary.BigEndian.Uint32(sum[sum[len(sum)-1]&0x0F:]) & 0x7FFFFFFF
	d := uint32(1)
	for i := 0; i < digits && i < 8; i++ {
		d *= 10
	}
	return int(v % d)
}
