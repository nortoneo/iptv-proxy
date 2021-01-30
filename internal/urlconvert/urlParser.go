package urlconvert

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/nortoneo/iptv-proxy/internal/config"
)

const (
	paramList      = "iptv_proxy_list"
	paramEncTarget = "iptv_proxy_target"
)

// GetParamList return list param key
func GetParamList() string {
	return paramList
}

// GetParamEncTarget return enc target param key
func GetParamEncTarget() string {
	return paramEncTarget
}

// ConvertPathToProxyPath convert path to proxy path by adding query params
func ConvertPathToProxyPath(path, listName, encURL string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set(GetParamList(), listName)
	q.Set(GetParamEncTarget(), encURL)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// ConvertURLtoProxyURL converts real url to proxy url
func ConvertURLtoProxyURL(realURL, appURL, listName string) (string, error) {
	real, err := url.Parse(realURL)
	if err != nil {
		return "", err
	}

	app, err := url.Parse(appURL)
	if err != nil {
		return "", err
	}

	//encoding real host path
	encURL := real.Scheme
	encURL += "://"
	if real.User.String() != "" {
		encURL += real.User.String() + "@"
	}
	encURL += real.Host

	key := config.GetConfig().EncryptionKey
	token, _ := config.GetListToken(listName)
	key += token

	encURL, err = Encode(encURL, key)
	if err != nil {
		return "", err
	}

	//overriding to proxy
	real.Scheme = app.Scheme
	real.Host = app.Host
	real.User = app.User
	q := real.Query()
	q.Set(GetParamList(), listName)
	q.Set(GetParamEncTarget(), encURL)
	real.RawQuery = q.Encode()

	proxyURLString := real.String()

	return proxyURLString, nil
}

// ConvertProxyRequestToURL converts request to target url string
func ConvertProxyRequestToURL(r *http.Request) (string, string, error) {
	appURL, err := url.Parse(config.GetConfig().AppURL)
	if err != nil {
		return "", "", err
	}

	reqURL := r.URL
	reqURL.Scheme = appURL.Scheme
	reqURL.Host = appURL.Host
	reqURL.User = appURL.User

	return ConvertProxyURLtoURL(reqURL.String())
}

// ConvertProxyURLtoURL converts real url to proxy url
// returns realURL, listName, error
func ConvertProxyURLtoURL(proxyURL string) (string, string, error) {
	pURL, err := url.Parse(proxyURL)
	if err != nil {
		return "", "", err
	}

	q := pURL.Query()
	listName := q.Get(paramList)
	if listName == "" {
		return "", "", errors.New("No list name provided")
	}
	encURL := q.Get(paramEncTarget)
	if encURL == "" {
		return "", "", errors.New("No target provided")
	}

	//removing proxy params
	q.Del(GetParamList())
	q.Del(GetParamEncTarget())
	pURL.RawQuery = q.Encode()

	key := config.GetConfig().EncryptionKey
	token, err := config.GetListToken(listName)
	if err != nil {
		return "", "", err
	}
	key += token

	decURL, err := Decode(encURL, key)
	if err != nil {
		return "", "", err
	}
	realURL, err := url.Parse(decURL)
	if err != nil {
		return "", "", err
	}

	pURL.Scheme = realURL.Scheme
	pURL.Host = realURL.Host
	pURL.User = realURL.User

	return pURL.String(), listName, nil
}

// Encode encodes string to obfuscated url friendly string
func Encode(text, key string) (string, error) {
	encrypted, err := encrypt(text, key)
	if err != nil {
		return "", err
	}
	gziped, err := gzipString(encrypted)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(gziped))

	return encoded, nil
}

// Decode decodes obfuscated string.
func Decode(encoded, key string) (string, error) {
	decodedBytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	unGziped, err := unGzipString(string(decodedBytes))
	if err != nil {
		return "", err
	}
	decrypted, err := decrypt(unGziped, key)
	if err != nil {
		return "", err
	}

	return decrypted, nil
}

func encrypt(stringToEncrypt string, keyString string) (string, error) {
	// todo add proper key derivation
	keySum := md5.Sum([]byte(keyString))
	keyString = hex.EncodeToString(keySum[:])

	//Since the key is in string, we need to convert decode it to bytes
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return "", nil
	}
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", nil
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return fmt.Sprintf("%x", ciphertext), nil
}

func decrypt(encryptedString string, keyString string) (string, error) {
	// todo add proper key derivation
	keySum := md5.Sum([]byte(keyString))
	keyString = hex.EncodeToString(keySum[:])

	key, err := hex.DecodeString(keyString)
	if err != nil {
		return "", err
	}
	enc, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}
	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	//Get the nonce size
	nonceSize := aesGCM.NonceSize()
	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", plaintext), nil
}

func gzipString(text string) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(text)); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	return string(b.Bytes()), nil
}

func unGzipString(text string) (string, error) {
	b := bytes.NewBufferString(text)
	gr, err := gzip.NewReader(b)
	if err != nil {
		return "", err
	}
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
