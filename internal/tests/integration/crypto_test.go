//+build integration

package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"gopkg.in/check.v1"

	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type CryptoSuite struct {
	tmpDir *fs.Dir
}

var _ = check.Suite(&CryptoSuite{})

func genTestKeys() (string, string) {
	rsaPriv, _ := rsa.GenerateKey(rand.Reader, 4096)
	rsaPub := rsaPriv.PublicKey
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaPriv),
	}
	testPrivKey := string(pem.EncodeToMemory(privBlock))

	b, _ := x509.MarshalPKIXPublicKey(&rsaPub)
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	testPubKey := string(pem.EncodeToMemory(pubBlock))
	return testPrivKey, testPubKey
}

func (s *CryptoSuite) SetUpTest(c *check.C) {
	testPrivKey, testPubKey := genTestKeys()
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("testPrivKey", testPrivKey),
		fs.WithFile("testPubKey", testPubKey),
	)
}

func (s *CryptoSuite) TearDownTest(c *check.C) {
	s.tmpDir.Remove()
}

func (s *CryptoSuite) TestRSACrypt(c *check.C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{ crypto.RSAGenerateKey 2048 -}}`,
		"-o", `key.pem`), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "privKey=./key.pem",
		"-i", `{{ $pub := crypto.RSADerivePublicKey .privKey -}}
{{ $enc := "hello" | crypto.RSAEncrypt $pub -}}
{{ crypto.RSADecryptBytes .privKey $enc | conv.ToString }}
{{ crypto.RSADecrypt .privKey $enc }}
`), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello\nhello\n"})
}
