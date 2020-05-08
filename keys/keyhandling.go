package keys

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sourcekris/goRsaTool/ln"
	fmp "github.com/sourcekris/goflint"
	"github.com/sourcekris/x509big"
)

var (
	// lineRE is a regexp that should match interesting integers on lines.
	lineRE = regexp.MustCompile(`(?i)^([n|e|c])\s*[:|=]\s*((?:0x)?[0-9a-f]+)$`)
	// numRE matches numbers in base 10 or hex.
	numRE = regexp.MustCompile(`[0-9a-f]+`)
	// modRE, expRE, ctRE matches 'n', 'e', 'c' case insensitively.
	modRE = regexp.MustCompile(`(?i)n`)
	expRE = regexp.MustCompile(`(?i)e`)
	ctRE  = regexp.MustCompile(`(?i)c`)
)

type pkParser func([]byte) (*x509big.BigPublicKey, error)

// parsePublicRsaKey attempts to try parsing the given public key yielding a FMPPublicKey or
// an error using multiple methods.
func parsePublicRsaKey(keyBytes []byte) (*FMPPublicKey, error) {
	var (
		parsers = []pkParser{
			x509big.ParseBigPKCS1PublicKey,
			x509big.ParseBigPKIXPublicKey,
		}
		errs []error
	)

	for _, p := range parsers {
		if key, err := p(keyBytes); err != nil {
			errs = append(errs, err)
		} else {
			return &FMPPublicKey{
				N: new(fmp.Fmpz).SetBytes(key.N.Bytes()),
				E: new(fmp.Fmpz).SetBytes(key.E.Bytes()),
			}, nil
		}
	}

	return nil, fmt.Errorf("parsePublicRsaKey failed: %v", errs)
}

func parsePrivateRsaKey(keyBytes []byte) (*FMPPrivateKey, error) {
	key, err := x509.ParsePKCS1PrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parsePrivateRsaKey: failed to parse the DER key after decoding: %v", err)
	}
	k := RSAtoFMPPrivateKey(key)
	return &k, nil
}

func parseBigPrivateRsaKey(keyBytes []byte) (*FMPPrivateKey, error) {
	key, err := x509big.ParseBigPKCS1PrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parseBigPrivateRsaKey: failed to parse the DER key after decoding: %v", err)
	}
	k := BigtoFMPPrivateKey(key)
	return &k, nil
}

// PrivateFromPublic takes a Public Key and return a Private Key with the public components packed.
func PrivateFromPublic(key *FMPPublicKey) *FMPPrivateKey {
	return &FMPPrivateKey{
		PublicKey: key,
		N:         key.N,
	}
}

// getBase returns the base of a string and, if its prefixed with 0x then the remainder of the string after the prefix.
func getBase(s string) (string, int) {
	if strings.HasPrefix(s, "0x") {
		return s[2:], 16
	}

	return s, 10
}

// ImportIntegerList attempts to parse the key (and optionally ciphertext) data as if it was a list of integers N, and e and c.
func ImportIntegerList(kb []byte) (*RSA, error) {
	var (
		n, e, c string
		ct      []byte
	)

	s := bufio.NewScanner(bytes.NewReader(kb))
	for s.Scan() {
		if lineRE.MatchString(s.Text()) {
			for _, sm := range lineRE.FindAllStringSubmatch(s.Text(), -1) {
				if len(sm) < 3 {
					continue
				}

				switch {
				case modRE.MatchString(sm[1]) && numRE.MatchString(sm[2]):
					n = sm[2]
				case expRE.MatchString(sm[1]) && numRE.MatchString(sm[2]):
					e = sm[2]
				case ctRE.MatchString(sm[1]) && numRE.MatchString(sm[2]):
					c = sm[2]
				}
			}
		}
	}

	if n == "" || e == "" {
		return nil, fmt.Errorf("failed to decode key, missing a modulus or an exponent")
	}

	fN, ok := new(fmp.Fmpz).SetString(getBase(n))
	if !ok {
		return nil, fmt.Errorf("failed decoding modulus from keyfile: %v", n)
	}

	fE, ok := new(fmp.Fmpz).SetString(getBase(e))
	if !ok {
		return nil, errors.New("failed decoding exponent from keyfile")
	}

	if c != "" {
		fC, ok := new(fmp.Fmpz).SetString(getBase(c))
		if !ok {
			return nil, errors.New("failed converting ciphertext integer to binary")
		}

		ct = ln.NumberToBytes(fC)
	}

	k, err := NewRSA(PrivateFromPublic(&FMPPublicKey{N: fN, E: fE}), ct, nil, "", false)
	if err != nil {
		return nil, err
	}

	return k, nil
}

// ImportKey imports a PEM key file and returns a FMPPrivateKey object or error.
func ImportKey(kb []byte) (*FMPPrivateKey, error) {
	// Decode the PEM data to extract the DER format key.
	block, _ := pem.Decode(kb)
	if block == nil {
		return nil, errors.New("failed to decode PEM key")
	}

	// Extract a FMPPublicKey from the DER decoded data and pack a private key struct.
	key, err := parsePublicRsaKey(block.Bytes)
	if err == nil {
		// If there was an error, try to parse it an alternative way below.
		return PrivateFromPublic(key), nil
	}

	priv, err := parseBigPrivateRsaKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ImportKey: failed to parse the key as either a public or private key: %v", err)
	}

	return priv, nil
}
