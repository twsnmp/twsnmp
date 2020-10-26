package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"golang.org/x/crypto/ssh"
)

// KeyPass : 秘密鍵のパスワード
const KeyPass = "TWSNMPba98be2110e9653f249aa2b38706cb02YMI"

// CertTerm : 自己署名の期限
const CertTerm = 1

func openStrURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		astiLogger.Errorf("openUrl err=%v", err)
		return err
	}
	return err
}

func openURL(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) < 1 {
		astiLogger.Errorf("openUrl no payload")
		return "ng", nil
	}
	var url string
	if err := json.Unmarshal(m.Payload, &url); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	if err := openStrURL(url); err != nil {
		return "ng", err
	}
	return "ok", nil
}

// genPrivateKey : Generate RSA Key
func genPrivateKey(bits int) (string, error) {
	// Generate the key of length bits
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", err
	}
	// Convert it to pem
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	// Encrypt the pem
	block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(KeyPass), x509.PEMCipherAES256)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(block)), nil
}

func getPEMBlocks(b []byte) ([]*pem.Block, error) {
	var ret []*pem.Block
	for len(b) > 0 {
		block, r := pem.Decode(b)
		if block == nil {
			return ret, fmt.Errorf("PEM Decode Error")
		}
		ret = append(ret, block)
		b = r
	}
	return ret, nil
}

func getRSAKeyFromPEMBlocks(blocks []*pem.Block) (*rsa.PrivateKey, error) {
	for _, block := range blocks {
		if block.Type == "RSA PRIVATE KEY" {
			if x509.IsEncryptedPEMBlock(block) {
				kder, e := x509.DecryptPEMBlock(block, []byte(KeyPass))
				if e == nil {
					return x509.ParsePKCS1PrivateKey(kder)
				}
			} else {
				return x509.ParsePKCS1PrivateKey(block.Bytes)
			}
		} else if block.Type == "PRIVATE KEY" {
			var b []byte
			var err error
			if x509.IsEncryptedPEMBlock(block) {
				b, err = x509.DecryptPEMBlock(block, []byte(KeyPass))
				if err != nil {
					return nil, err
				}
			} else {
				b = block.Bytes
			}
			keyInterface, err := x509.ParsePKCS8PrivateKey(b)
			if err != nil {
				return nil, err
			}
			key, ok := keyInterface.(*rsa.PrivateKey)
			if ok {
				return key, nil
			}
		}
	}
	return nil, fmt.Errorf("RSA key not found in pem")
}

// getRSAKeyFromPEM : 暗号化されたPEMデータから秘密鍵を取得する
func getRSAKeyFromPEM(p string) (*rsa.PrivateKey, error) {
	blocks, err := getPEMBlocks([]byte(p))
	if err != nil {
		return nil, err
	}
	return getRSAKeyFromPEMBlocks(blocks)
}

func getCnAlt() (string, string) {
	host, err := os.Hostname()
	if err != nil {
		astiLogger.Errorf("getCnAlt err=%v", err)
		return "TWSNMP", "TWSNMP"
	}
	alts := []string{host}
	ifs, err := net.Interfaces()
	if err != nil {
		astiLogger.Errorf("getCnAlt err=%v", err)
		return "TWSNMP", "TWSNMP"
	}
	for _, i := range ifs {
		if (i.Flags&net.FlagLoopback) == net.FlagLoopback ||
			(i.Flags&net.FlagUp) != net.FlagUp {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			cidr := a.String()
			ip, _, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			if ip.To4() == nil {
				continue
			}
			alts = append(alts, ip.String())
		}
	}
	return host, strings.Join(alts, ",")
}

// genSelfSignCert : 自己署名の証明書を作成する
func genSelfSignCert(key string) (string, error) {
	cn, alt := getCnAlt()
	subj := pkix.Name{
		CommonName: cn,
		// Country:            []string{Config.CsrCountry},
		// Province:           []string{Config.CsrProvince},
		// Locality:           []string{Config.CsrLocality},
		// Organization:       []string{Config.CsrOrg},
		// OrganizationalUnit: []string{Config.CsrOrgUnit},
	}
	keyBytes, err := getRSAKeyFromPEM(key)
	if err != nil {
		return "", err
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(CertTerm * 365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", fmt.Errorf("failed to generate serial number: %s", err)
	}
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subj,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	hosts := strings.Split(alt, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &keyBytes.PublicKey, keyBytes)
	if err != nil {
		return "", fmt.Errorf("Failed to create certificate: %s", err)
	}
	cert := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))
	return cert, nil
}

// getSSHPublicKey : SSHの公開鍵をOpenSSH形式で取得する
func getSSHPublicKey(key string) (string, error) {
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	comment := fmt.Sprintf("twsnmp@%s", host)
	priv, err := getRSAKeyFromPEM(key)
	if err != nil {
		return "", fmt.Errorf("Key Not Found")
	}
	rsaKey := priv.PublicKey
	pubkey, _ := ssh.NewPublicKey(&rsaKey)
	return fmt.Sprintf("%s %s", strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pubkey))), comment), nil
}

// getRawKeyPem : 暗号化を解除した秘密鍵のPEMを取得する
func getRawKeyPem(p string) string {
	priv, err := getRSAKeyFromPEM(p)
	if err != nil {
		return ""
	}
	// Convert it to pem
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}
	return string(pem.EncodeToMemory(block))
}
