package util

import (
	"code.google.com/p/gorilla/securecookie"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
	//	"net"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
)

var authName string = "authentication.key"
var encName string = "encryption.key"

func CookieKeys() (auth, enc []byte) {
	auth, enc = cookieKey(authName), cookieKey(encName)
	return
}

func cookieKey(fname string) (data []byte) {
	var err error
	data, err = ioutil.ReadFile(filepath.Join(BaseDir(), fname))
	if err != nil {
		data = securecookie.GenerateRandomKey(32)
		err = SaveFile(filepath.Join(BaseDir(), fname), data)
		if err != nil {
			Log(err)
		}
	}
	return data
}

//Validate authenticates a provided password against a hashed password.
func Validate(hashed, salt, pword string) bool {
	computed := ComputeHash(pword, salt)
	return hashed == computed
}

//Hash hashes the provided password.
func Hash(pword string) (hash, salt string) {
	salt = GenString(32)
	return ComputeHash(pword, salt), salt
}

//ComputeHash computes the hash for a password and its salt.
func ComputeHash(pword, salt string) string {
	h := sha1.New()
	io.WriteString(h, pword+salt)
	return hex.EncodeToString(h.Sum(nil))
}

//GenString generates a psuedo-random string using the crypto/rand package.
func GenString(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	en := base64.StdEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}

func GenCertificate(certName, keyName string) (err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return
	}
	notBefore := time.Now()
	notAfter := time.Date(2049, 12, 31, 23, 59, 59, 0, time.UTC)
	template := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(1),
		Subject: pkix.Name{
			Organization: []string{"Stellenbosch University"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:         true,
		SubjectKeyId: []byte{1, 2, 3, 4},
		Version:      2,
	} 
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return
	}
	certOut, err := os.OpenFile(certName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer certOut.Close()
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return
	}
	keyOut, err := os.OpenFile(keyName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer keyOut.Close()
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		return
	}
	return
}
