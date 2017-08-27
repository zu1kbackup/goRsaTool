package attacks

import (
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"
  "errors"
  "fmt"
  "math"
  "math/big"
  "github.com/ncw/gmp"
  "github.com/sourcekris/goRsaTool/libnum"
  "github.com/sourcekris/goRsaTool/utils"
)

// final internal representation for keys
type GMPPublicKey struct {
  N *gmp.Int
  E *gmp.Int
}

type GMPPrivateKey struct {
  PublicKey *GMPPublicKey
  D *gmp.Int
  Primes []*gmp.Int
  N *gmp.Int
}

// an unfortunate intermediate representation for keys
type BigPublicKey struct {
  N *big.Int
  E *big.Int
}

type BigPrivateKey struct {
  PublicKey *BigPublicKey
  D *big.Int
  Primes []*big.Int
  N *big.Int
}

/*
 * wrap rsa.PrivateKey and add a field for cipher and plaintexts
 */
type RSAStuff struct {
  Key GMPPrivateKey
  CipherText []byte
  PlainText []byte
  PastPrimesFile string
}

/*
 * constructor for RSAStuff struct
 */
func NewRSAStuff(key *BigPrivateKey, c []byte, m []byte, pf string) (*RSAStuff, error) {
	if key.PublicKey.N == nil {
		return nil, errors.New("Key had no modulus or exponent")
	}

  var pastPrimesFile string
    if len(pf) > 0 {
      pastPrimesFile = pf
  }

  var cipherText []byte
    if len(c) > 0 {
      cipherText = c
  }

  // copy a rsa.PrivateKey to a GMPPrivateKey that uses gmp.Int types
  gmpPrivateKey := BigtoGMPPrivateKey(key)

	// pack the RSAStuff struct
   return &RSAStuff{
      Key: gmpPrivateKey,
      PastPrimesFile: pastPrimesFile,
      CipherText: cipherText,
    }, nil
}

/*
 * Given one prime p, pack the Key member of the RSAStuff struct with the private key values, p, q & d
 */
func (targetRSA *RSAStuff) PackGivenP(p *gmp.Int) {
  q := new(gmp.Int).Div(targetRSA.Key.N, p)
  targetRSA.Key.Primes = []*gmp.Int{p, q}
  targetRSA.Key.D      = utils.SolveforD(p, q, targetRSA.Key.PublicKey.E)
}

func (targetRSA *RSAStuff) DumpKey() {
  fmt.Printf("[*] n = %d\n", targetRSA.Key.PublicKey.N)
  fmt.Printf("[*] e = %d\n", targetRSA.Key.PublicKey.E)

  // XXX: Support RSA multiprime [where len(key.Primes) > 2]
  if targetRSA.Key.D!= nil {
    fmt.Printf("[*] d = %d\n", targetRSA.Key.D)
    fmt.Printf("[*] p = %d\n", targetRSA.Key.Primes[0])
    fmt.Printf("[*] q = %d\n", targetRSA.Key.Primes[1])
  }

  if len(targetRSA.CipherText) > 0 {
    fmt.Printf("[*] c = %d\n", libnum.BytesToNumber(targetRSA.CipherText))
  }
}

/*
 * Takes a rsa.PrivateKey and returns a GMPPrivateKey that uses gmp.Int types
 */
func RSAtoGMPPrivateKey(key *rsa.PrivateKey) GMPPrivateKey {
  gmpPubKey := &GMPPublicKey{
    N: new(gmp.Int).SetBytes(key.N.Bytes()),
    E: gmp.NewInt(int64(key.E)),
  }

  var gmpPrivateKey *GMPPrivateKey
  if key.D != nil {
    gmpPrivateKey = &GMPPrivateKey{
      PublicKey: gmpPubKey,
      D: new(gmp.Int).SetBytes(key.D.Bytes()),
      Primes: []*gmp.Int{
        new(gmp.Int).SetBytes(key.Primes[0].Bytes()), 
        new(gmp.Int).SetBytes(key.Primes[1].Bytes()),
        },
    }
  } else {
    gmpPrivateKey = &GMPPrivateKey{
      PublicKey: gmpPubKey,
      N: new(gmp.Int).SetBytes(key.N.Bytes()),
    }
  }

  return *gmpPrivateKey
}

func GMPtoRSAPrivateKey(key *GMPPrivateKey) *rsa.PrivateKey {
  if key.PublicKey.E.Cmp(gmp.NewInt(math.MaxInt64)) > 0 {
    // XXX todo: handle better? phase out rsa.PrivateKey types
    panic("[-] Exponent is too large for the private key to be converted to type rsa.PrivateKey")
    
  }

  pubKey := &rsa.PublicKey{
    N: new(big.Int).SetBytes(key.N.Bytes()),
    E: int(key.PublicKey.E.Int64()),
  }

  var privateKey *rsa.PrivateKey
  if key.D != nil {
    privateKey = &rsa.PrivateKey{
      PublicKey: *pubKey,
      D: new(big.Int).SetBytes(key.D.Bytes()),
      Primes: []*big.Int{
        new(big.Int).SetBytes(key.Primes[0].Bytes()), 
        new(big.Int).SetBytes(key.Primes[1].Bytes()),
      },
    }    
  } else {
    privateKey = &rsa.PrivateKey{
      PublicKey: *pubKey,
    }
  }

  return privateKey
}

func RSAtoBigPrivateKey(key *rsa.PrivateKey) BigPrivateKey {
  bigPubKey := &BigPublicKey{
    N: new(big.Int).SetBytes(key.N.Bytes()),
    E: big.NewInt(int64(key.E)),
  }

  var bigPrivateKey *BigPrivateKey
  if key.D != nil {
    bigPrivateKey = &BigPrivateKey{
      PublicKey: bigPubKey,
      D: new(big.Int).SetBytes(key.D.Bytes()),
      Primes: []*big.Int{
        new(big.Int).SetBytes(key.Primes[0].Bytes()), 
        new(big.Int).SetBytes(key.Primes[1].Bytes()),
        },
    }
  } else {
    bigPrivateKey = &BigPrivateKey{
      PublicKey: bigPubKey,
      N: new(big.Int).SetBytes(key.N.Bytes()),
    }
  }

  return *bigPrivateKey
}

func BigtoRSAPrivateKey(key *BigPrivateKey) *rsa.PrivateKey {
  if key.PublicKey.E.Cmp(big.NewInt(math.MaxInt64)) > 0 {
    // XXX todo: handle better? phase out rsa.PrivateKey types
    panic("[-] Exponent is too large for the private key to be converted to type rsa.PrivateKey")
    
  }

  pubKey := &rsa.PublicKey{
    N: new(big.Int).SetBytes(key.N.Bytes()),
    E: int(key.PublicKey.E.Int64()),
  }

  var privateKey *rsa.PrivateKey
  if key.D != nil {
    privateKey = &rsa.PrivateKey{
      PublicKey: *pubKey,
      D: new(big.Int).SetBytes(key.D.Bytes()),
      Primes: []*big.Int{
        new(big.Int).SetBytes(key.Primes[0].Bytes()), 
        new(big.Int).SetBytes(key.Primes[1].Bytes()),
      },
    }    
  } else {
    privateKey = &rsa.PrivateKey{
      PublicKey: *pubKey,
    }
  }

  return privateKey
}

func BigtoGMPPrivateKey(key *BigPrivateKey) GMPPrivateKey {
  gmpPubKey := &GMPPublicKey{
    N: new(gmp.Int).SetBytes(key.PublicKey.N.Bytes()),
    E: new(gmp.Int).SetBytes(key.PublicKey.E.Bytes()),
  }

  var gmpPrivateKey *GMPPrivateKey
  if key.D != nil {
      
    gmpPrivateKey = &GMPPrivateKey{
      PublicKey: gmpPubKey,
      D: new(gmp.Int).SetBytes(key.D.Bytes()),
      Primes: []*gmp.Int{
        new(gmp.Int).SetBytes(key.Primes[0].Bytes()), 
        new(gmp.Int).SetBytes(key.Primes[1].Bytes()),
        },
    }
  } else {
    gmpPrivateKey = &GMPPrivateKey{
      PublicKey: gmpPubKey,
      N: new(gmp.Int).SetBytes(key.PublicKey.N.Bytes()),
    }
  }

  return *gmpPrivateKey
}

func encodeDerToPem(der []byte, t string) string {
  p := pem.EncodeToMemory(
    &pem.Block{
      Type: t, 
      Bytes: der,
      },
      )

  return string(p)
}

func EncodePublicKey(pub *rsa.PublicKey) (string, error) {
  pubder, err := x509.MarshalPKIXPublicKey(pub)
  if err != nil {
    return "", err
  }

  return encodeDerToPem(pubder, "RSA PUBLIC KEY"), nil
}

func EncodePrivateKey(priv *rsa.PrivateKey) string {
  privder := x509.MarshalPKCS1PrivateKey(priv)
  return encodeDerToPem(privder, "RSA PRIVATE KEY")
}

func EncodeGMPPrivateKey(priv *GMPPrivateKey) string {
  privder := x509.MarshalPKCS1PrivateKey(GMPtoRSAPrivateKey(priv))
  return encodeDerToPem(privder, "RSA PRIVATE KEY")
}