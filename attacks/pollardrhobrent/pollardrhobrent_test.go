package pollardrhobrent

import (
	"testing"

	"github.com/sourcekris/goRsaTool/keys"
	"github.com/sourcekris/goRsaTool/ln"
	fmp "github.com/sourcekris/goflint"
)

func foundP(p *fmp.Fmpz, ps []*fmp.Fmpz) bool {
	for _, prime := range ps {
		if p.Cmp(prime) == 0 {
			return true
		}
	}
	return false
}

func TestAttack(t *testing.T) {
	tt := []struct {
		name    string
		n       *fmp.Fmpz
		wantP   *fmp.Fmpz
		wantErr bool
	}{
		{
			name:  "vulnerable key expected to factor",
			n:     ln.FmpString("115792089237316195423570985008687907853269984665640564039457584007913129639937"),
			wantP: ln.FmpString("1238926361552897"),
		},
	}

	for _, tc := range tt {
		fmpPubKey := &keys.FMPPublicKey{
			N: tc.n,
			E: fmp.NewFmpz(65537),
		}

		k, _ := keys.NewRSA(keys.PrivateFromPublic(fmpPubKey), nil, nil, "", false)
		err := Attack([]*keys.RSA{k})
		if err != nil && !tc.wantErr {
			t.Errorf("Attack() failed: %s expected no error got error: %v", tc.name, err)
		}

		if k.Key.D == nil && !tc.wantErr {
			t.Errorf("Attack() failed: %s d not found", tc.name)
		}

		if !foundP(tc.wantP, k.Key.Primes) {
			t.Errorf("Attack() failed: %s expected primes not found - got %v wanted %v", tc.name, k.Key.Primes, tc.wantP)
		}
	}
}
