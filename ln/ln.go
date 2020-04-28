package ln

import (
	"github.com/jbarham/primegen"
	"github.com/kavehmz/prime"

	fmp "github.com/sourcekris/goflint"
)

const (
	// The number of Miller-Rabin rounds for Golangs ProbablyPrime.
	mrRounds = 20
)

var (
	// BigNOne is the Fmpz representation of -1
	BigNOne = fmp.NewFmpz(-1)
	// BigZero is the Fmpz representation of 0
	BigZero = fmp.NewFmpz(0)
	//  is the Fmpz representation of 1
	BigOne = fmp.NewFmpz(1)
	// BigTwo is the Fmpz representation of 2
	BigTwo = fmp.NewFmpz(2)
	// BigThree is the Fmpz representation of 3
	BigThree = fmp.NewFmpz(3)
	// BigFour is the Fmpz representation of 4
	BigFour = fmp.NewFmpz(4)
	// BigFive is the Fmpz representation of 5
	BigFive = fmp.NewFmpz(5)
	// BigSix is the Fmpz representation of 6
	BigSix = fmp.NewFmpz(6)
	// BigSeven is the Fmpz representation of 7
	BigSeven = fmp.NewFmpz(7)
	// BigEight is the Fmpz representation of 8
	BigEight = fmp.NewFmpz(8)
	// BigNine is the Fmpz representation of 9
	BigNine = fmp.NewFmpz(9)
	// BigEleven is the Fmpz representation of 11
	BigEleven = fmp.NewFmpz(11)
	// BigSixteen is the Fmpz representation of 16
	BigSixteen = fmp.NewFmpz(0xf)
)

// BytesToNumber takes a slice of bytes and returns a Fmpz integer representation.
func BytesToNumber(src []byte) *fmp.Fmpz {
	return new(fmp.Fmpz).SetBytes(src)
}

// NumberToBytes takes an Fmpz integer and returns the byte slice representation.
func NumberToBytes(src *fmp.Fmpz) []byte {
	return src.Bytes()
}

// SolveforD given e, p and q solve for the private exponent d.
func SolveforD(p *fmp.Fmpz, q *fmp.Fmpz, e *fmp.Fmpz) *fmp.Fmpz {
	return new(fmp.Fmpz).ModInverse(e,
		new(fmp.Fmpz).Mul(
			new(fmp.Fmpz).Sub(p, BigOne),
			new(fmp.Fmpz).Sub(q, BigOne),
		),
	)
}

// FindPGivenD finds p and q given d, e, and n - uses an algorithm from pycrypto _slowmath.py [0]
// [0]: https://github.com/dlitz/pycrypto/blob/master/lib/Crypto/PublicKey/_slowmath.py
func FindPGivenD(d *fmp.Fmpz, e *fmp.Fmpz, n *fmp.Fmpz) *fmp.Fmpz {
	m := new(fmp.Fmpz)
	tmp := new(fmp.Fmpz)

	ktot := new(fmp.Fmpz).Set(tmp.Mul(d, e).Sub(tmp, BigOne))
	t := new(fmp.Fmpz).Set(ktot)

	for tmp.Mod(t, BigTwo).Cmp(BigZero) == 0 {
		t.DivMod(t, BigTwo, m)
	}

	for a := 2; a < 1000; a += 2 {
		k := new(fmp.Fmpz).Set(t)

		cand := new(fmp.Fmpz)
		for k.Cmp(ktot) < 0 {
			cand.Exp(fmp.NewFmpz(int64(a)), k, n)

			if cand.Cmp(BigOne) != 0 && cand.Cmp(tmp.Sub(n, BigOne)) != 0 && tmp.Exp(cand, BigTwo, n).Cmp(BigOne) == 0 {
				return FindGcd(tmp.Add(cand, BigOne), n)
			}

			k.Mul(k, BigTwo)
		}
	}

	return BigZero
}

// IsPerfectSquare returns t if n is a perfect square -1 otherwise
func IsPerfectSquare(n *fmp.Fmpz) *fmp.Fmpz {
	h := new(fmp.Fmpz).And(n, BigSixteen)

	if h.Cmp(fmp.NewFmpz(9)) > 1 {
		return BigNOne
	}

	if h.Cmp(BigTwo) != 0 && h.Cmp(BigThree) != 0 &&
		h.Cmp(BigFive) != 0 && h.Cmp(BigSix) != 0 &&
		h.Cmp(BigSeven) != 0 && h.Cmp(BigEight) != 0 {

		t := new(fmp.Fmpz).Sqrt(n)

		if t.Mul(t, t).Cmp(n) == 0 {
			return t
		} else {
			return BigNOne
		}
	}

	return BigNOne
}

// RationalToContfract takes a rational represented by x and y and returns a slice of quotients.
func RationalToContfract(x, y *fmp.Fmpz) []int {
	a := new(fmp.Fmpz).Div(x, y)
	b := new(fmp.Fmpz).Mul(a, y)
	c := new(fmp.Fmpz)

	var pquotients []int

	if b.Cmp(x) == 0 {
		return []int{int(a.Int64())}
	}
	c.Mul(y, a).Sub(x, c)
	pquotients = RationalToContfract(y, c)
	pquotients = append([]int{int(a.Int64())}, pquotients...)
	return pquotients
}

// ContfractToRational takes a slice of quotients and returns a rational x/y.
func ContfractToRational(frac []int) (*fmp.Fmpz, *fmp.Fmpz) {
	var remainder []int

	switch l := len(frac); l {
	case 0:
		return BigZero, BigOne
	case 1:
		return fmp.NewFmpz(int64(frac[0])), BigOne
	default:
		remainder = frac[1:l]
		num, denom := ContfractToRational(remainder)
		fracZ := fmp.NewFmpz(int64(frac[0]))
		return fracZ.Mul(fracZ, num).Add(fracZ, denom), num
	}
}

// ConvergantsFromContfract takes a slice of quotients and returns the convergants.
func ConvergantsFromContfract(frac []int) [][2]*fmp.Fmpz {
	var convs [][2]*fmp.Fmpz

	for i := range frac {
		a, b := ContfractToRational(frac[0:i])
		z := [2]*fmp.Fmpz{a, b}
		convs = append(convs, z)
	}
	return convs
}

// SieveOfEratosthenes returns primes from begin to n and this implementation comes from:
// https://stackoverflow.com/a/21923233.
func SieveOfEratosthenes(n int) []int {
	var primes []int
	b := make([]bool, n)
	for i := 2; i < n; i++ {
		if b[i] == true {
			continue
		}

		primes = append(primes, i)
		for k := i * i; k < n; k += i {
			b[k] = true
		}
	}
	return primes
}

// SieveOfEratosthenesFmp is a convenience function that simply returns []*fmp.Fmpz instead of
// []int. It does not support finding primes > max_int.
func SieveOfEratosthenesFmp(n int) []*fmp.Fmpz {
	return FmpFromIntSlice(SieveOfEratosthenes(n))
}

// SieveRangeOfAtkin finds primes from begin to the limit n using a SieveOfAtkin from primegen package.
func SieveRangeOfAtkin(begin, n int) []int {
	var primes []int

	pg := primegen.New()
	pg.SkipTo(uint64(begin))
	for i := begin; i < n; i++ {
		primes = append(primes, int(pg.Next()))
	}
	return primes
}

// SieveOfAtkin finds primes from 0 to n using a SieveOfAtkin from primegen package.
func SieveOfAtkin(n int) []int {
	return SieveRangeOfAtkin(2, n)
}

// SieveOfAtkinFmp finds primes from 0 to n using a SieveOfAtkin from primegen package.
func SieveOfAtkinFmp(n int) []*fmp.Fmpz {
	return FmpFromIntSlice(SieveRangeOfAtkin(2, n))
}

// SegmentedSieveFmp is another prime sieve.
func SegmentedSieveFmp(n int) []*fmp.Fmpz {
	return FmpFromUInt64Slice(prime.Primes(uint64(n)))
}

// FmpFromIntSlice returns a slice of Fmpz from a slice of int.
func FmpFromIntSlice(is []int) []*fmp.Fmpz {
	var res []*fmp.Fmpz
	for _, i := range is {
		res = append(res, fmp.NewFmpz(int64(i)))
	}
	return res
}

// FmpFromUInt64Slice returns a slice of Fmpz from a slice of int.
func FmpFromUInt64Slice(is []uint64) []*fmp.Fmpz {
	var res []*fmp.Fmpz
	for _, i := range is {
		res = append(res, new(fmp.Fmpz).SetUint64(i))
	}
	return res
}

// FmpString returns a base 10 fmp.Fmpz integer from a string. It returns BigZero on error.
func FmpString(s string) *fmp.Fmpz {
	res, ok := new(fmp.Fmpz).SetString(s, 10)
	if !ok {
		return BigZero
	}

	return res
}

// MLucas multiplies along a Lucas sequence modulo n.
func MLucas(v, a, n *fmp.Fmpz) *fmp.Fmpz {
	v1 := new(fmp.Fmpz).Set(v)
	v2 := new(fmp.Fmpz).Mul(v, v)
	v2.Sub(v2, BigTwo).Mod(v2, n)

	for i := a.Bits(); i > 1; i-- {
		if a.TstBit(i) == 0 {
			tmpv1 := new(fmp.Fmpz).Set(v1)
			v1.Mul(v1, v1).Sub(v1, BigTwo).Mod(v1, n)
			v2.Mul(tmpv1, v2).Sub(v2, v).Mod(v2, n)
		} else {
			v1.Mul(v1, v2).Sub(v1, v).Mod(v1, n)
			v2.Mul(v2, v2).Sub(v2, BigTwo).Mod(v2, n)
		}
	}

	return v1
}
