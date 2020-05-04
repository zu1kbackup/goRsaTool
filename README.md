# goRsaTool

A golang port of [RsaCtfTool](https://github.com/sourcekris/RsaCtfTool) for the express purposes of learning go.

RsaCtfTool is an RSA tool for CTF challenges, it attempts multiple attacks against a public key in an effort to recover either the private key, the plain text of the message or both.

Attacks supported in this go version:

* factordb attack (i.e. is the modulus already fully factored on factordb.com)
* small q attack
* novelty primes attack
* past CTF primes attack
* fermat factorization for close p & q
* low public exponent attack (requires ciphertext)
* wiener's attack for large public exponents
* pollards p-1 attack
* williams p+1 attack
* pollards rho factorization - original Pollard's Monte Carlo factorization method
* pollard rho brent factorization - Richard Brents improved version of Pollard's monte carlo factorization
* Qi Cheng factorization from "A New Class of Unsafe Primes"

## Installation
 * Requires go 1.9 +
 * Get dependencies, you will need:
    * Golang
    * Git
    * FLINT2 (Fast Library for Number Theory)
    * This works on Debian and Ubuntu (tested 20.04 LTS)
    ```
    sudo apt install git golang libflint-dev
    ```
 * Get Golang libraries this uses:
   ```
   go get github.com/jbarham/primegen
   go get github.com/sourcekris/goflint
   go get github.com/sourcekris/x509big
   ```
 * Download and build this tool using go install repo and build:
   ```
   $ go get github.com/sourcekris/goRsaTool
   $ go install github.com/sourcekris/goRsaTool
   $ $HOME/go/bin/goRsaTool -h
    Usage of /home/sewid/go/bin/goRsaTool:
      -attack string
          Specific attack to try. Specify "all" for everything that works unnatended. (default "all")
      -ciphertext string
          An RSA encrypted binary file to decrypt, necessary for certain attacks.
      -createkey
          Create a public key given an E and N.
      -ctlist string
          Comma seperated list of keys for multi-key attacks.
      -dumpkey
          Just dump the RSA integers from a key - n,e,d,p,q.
      -e string
          The exponent value - for use with createkey flag.
      -key string
          The filename of the RSA key to attack or dump
      -keylist string
          Comma seperated list of keys for multi-key attacks.
      -list
          List the attacks supported by the attack flag.
      -n string
          The modulus value - for use with createkey flag.
      -pastprimes string
          The filename of a file containing past CTF prime numbers. (default "../pastctfprimes.txt")
      -verbose
          Enable verbose output.
   ```
 

## Usage:

### Generate a public key:
`./gorsatool -createkey -n 7828374823761928712873129873981723...12837182 -e 65537`

### Dump the parameters from a key:
`./gorsatool -dumpkey -key ./key.pub`

### Attack a public key:
`./gorsatool -key ./key.pub -attacks all`

### List available attacks:
`./gorsatool -list`

### Attack a public key with a specific attack:
`./gorsatool -key ./key.pub -attacks wiener`

### Attack the example pollards p-1 key with the pollards p-1 attack:
`./gorsatool -key examples/pollardsp1.pub -attacks pollardsp1`