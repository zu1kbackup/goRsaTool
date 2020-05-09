package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sourcekris/goRsaTool/attacks"
	"github.com/sourcekris/goRsaTool/keys"
	"github.com/sourcekris/goRsaTool/utils"

	fmp "github.com/sourcekris/goflint"
)

var (
	keyFile        = flag.String("key", "", "The filename of the RSA key to attack or dump")
	pastPrimesFile = flag.String("pastprimes", "../pastctfprimes.txt", "The filename of a file containing past CTF prime numbers.")
	verboseMode    = flag.Bool("verbose", false, "Enable verbose output.")
	dumpKeyMode    = flag.Bool("dumpkey", false, "Just dump the RSA integers from a key - n,e,d,p,q.")
	createKeyMode  = flag.Bool("createkey", false, "Create a public key given an E and N.")
	exponentArg    = flag.String("e", "", "The exponent value - for use with createkey flag.")
	modulusArg     = flag.String("n", "", "The modulus value - for use with createkey flag.")
	cipherText     = flag.String("ciphertext", "", "An RSA encrypted binary file to decrypt, necessary for certain attacks.")
	keyList        = flag.String("keylist", "", "Comma seperated list of keys for multi-key attacks.")
	ctList         = flag.String("ctlist", "", "Comma seperated list of ciphertext binaries for multi-key attacks.")
	attack         = flag.String("attack", "all", "Specific attack to try. Specify \"all\" for everything that works unnatended.")
	list           = flag.Bool("list", false, "List the attacks supported by the attack flag.")
	logger         *log.Logger
)

// unnatended will run all supported attacks against t that are listed as working in unnatended mode.
func unnatended(t *keys.RSA) []error {
	var errs []error
	for _, a := range attacks.SupportedAttacks.Supported {
		if a.Unnatended {
			if err := attacks.SupportedAttacks.Execute(a.Name, t); err != nil {
				errs = append(errs, err)
			}
			if t.Key.D != nil {
				if t.Verbose {
					logger.Printf("key factored with attack: %v\n", a.Name)
				}
				return nil
			}
		}
	}

	return errs
}

// listAttacks returns a string containing the list of registered attacks.
func listAttacks() string {
	var res string
	for _, a := range attacks.SupportedAttacks.Supported {
		res = fmt.Sprintf("%s%s\n", res, a.Name)
	}

	return res
}

func main() {
	flag.Parse()

	logger = log.New(os.Stderr, "rsatool: ", log.Lshortfile)

	if *verboseMode {
		logger.Println("staring up...")
	}

	if *list {
		fmt.Print(listAttacks())
		return
	}

	// Handle multi key scenarios.
	switch {
	case len(*keyList) > 0 && len(*ctList) == 0:
		// TODO(sewid): Loop around the key list and also do multi key attacks.
	case len(*keyList) > 0 && len(*ctList) > 0:
		// TODO(sewid): hastads broadcast attack.
	}

	// Did we get a public key file to read
	if *keyFile != "" {
		var (
			err       error
			targetRSA *keys.RSA
			nonPemKey bool
		)

		kb, err := ioutil.ReadFile(*keyFile)
		if err != nil {
			log.Fatalf("failed to open key file %q: %v", keyFile, err)
		}

		key, err := keys.ImportKey(kb)
		if err != nil {
			// Failed to read a valid PEM key. Maybe it is an integer list type key?
			targetRSA, err = keys.ImportIntegerList(kb)
			if err != nil {
				logger.Fatalf("failed reading key file: %v", err)
			}

			nonPemKey = true
			targetRSA.PastPrimesFile = *pastPrimesFile
			targetRSA.Verbose = *verboseMode
		}

		var c []byte
		if len(*cipherText) > 0 {
			c, err = utils.ReadCipherText(*cipherText)
			if err != nil {
				logger.Fatalf("failed reading ciphertext file: %v", err)
			}
		}

		if targetRSA == nil {
			targetRSA, err = keys.NewRSA(key, c, nil, *pastPrimesFile, *verboseMode)
			if err != nil {
				log.Fatalf("failed to create a RSA key from given key data: %v", err)
			}
		}

		if *dumpKeyMode {
			targetRSA.DumpKey()

			if nonPemKey {
				// The input was an integer list key so the user might actually want a PEM dump.
				fmt.Println(keys.EncodeFMPPublicKey(targetRSA.Key.PublicKey))
			}

			return
		}

		var errs []error
		switch {
		case *attack == "all":
			errs = unnatended(targetRSA)
		case attacks.SupportedAttacks.IsSupported(*attack):
			errs = append(errs, attacks.SupportedAttacks.Execute(*attack, targetRSA))
		default:
			errs = []error{fmt.Errorf("unsupported attack: %v. Use -list to see a list of supported attacks", *attack)}
		}

		for _, e := range errs {
			if e != nil {
				logger.Println(e)
			}
		}

		// were we able to solve for the private key?
		if targetRSA.Key.D != nil {
			fmt.Println(keys.EncodeFMPPrivateKey(&targetRSA.Key))
			return
		}

		if len(targetRSA.PlainText) > 0 {
			fmt.Println("Recovered plaintext: ")
			fmt.Println(string(targetRSA.PlainText))
		}

	} else {
		if *createKeyMode {
			if len(*exponentArg) > 0 && len(*modulusArg) > 0 {
				n, ok := new(fmp.Fmpz).SetString(*modulusArg, 10)
				if !ok {
					logger.Fatalf("failed converting modulus to integer: %q", *modulusArg)
				}

				e, ok := new(fmp.Fmpz).SetString(*exponentArg, 10)
				if !ok {
					logger.Fatalf("failed converting exponent to integer: %q", *exponentArg)
				}

				pubStr := keys.EncodeFMPPublicKey(&keys.FMPPublicKey{N: n, E: e})
				fmt.Println(pubStr)
				return
			}
			logger.Fatal("no exponent or modulus specified - use -n and -e")
		}
		logger.Fatal("no key file specified - use the -key flag to provide a public or private key file")
	}
}