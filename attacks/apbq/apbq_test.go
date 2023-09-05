package apbq

import (
	"testing"

	"github.com/sourcekris/goRsaTool/keys"
	"github.com/sourcekris/goRsaTool/ln"
	fmp "github.com/sourcekris/goflint"
)

func TestAttack(t *testing.T) {
	tt := []struct {
		name    string
		n       *fmp.Fmpz
		e       *fmp.Fmpz
		ct      *fmp.Fmpz
		max     int64
		hints   []*fmp.Fmpz
		want    string
		wantErr bool
	}{
		{
			name: "valid test case",
			n:    ln.FmpString("13487244535121893803142050477818837867090773702695830915710317760278957239414594039413664548291850262812704115774527807319037549055454297206076220984691198037713266404171521885962954384144959347235389444100155877481802912357132674633884880128105667841540583748054023374707572496059441301607888647200707488850720006967106436804871202685875375533545360179923969238661369697669827308101918547610915038310318070624021040766421119809895329315396306786911716715244892126715656507342336911573357257410955954494465940402266123528623572966813645357903662041629905600305564019544745386629585429281789016899281488949804805973433"),
			e:    ln.FmpString("65537"),
			ct:   ln.FmpString("7925658536205496145496105864909913841698804988627111589327264207647087371021599624715146199970201133465829350522657974209302809912914631345754196951377499186210285843997712271596344624581015221675171875097569926177625803286344226123963846381574190015963241702836267717409375800964065380453319977184702630199380943887323208760590947005727571317068147150612752450492200509903330780828198170278507237646300390745422616530575815926105334351017776515320327803006039040793248236695404925877281545258818155971734055166797929677109873068535807756177152624750247758835508005818076202086557580467517459509526459954994222107733"),
			max:  256,
			hints: []*fmp.Fmpz{
				ln.FmpString("93690707048761378546891432612703094136123056947302469539537929609977103203297047979247035258430608394707452208616011425282532322585909723570657884371221308059003099931556771434286270777087304918068710314109719362812230577136184026842003856478431246529965153009860967402874474597095746752792361627432414860218876940868512361825848930925319484457710800935318644177626456242425726362235994549199312317555"),
				ln.FmpString("350764904379382307689364277345531820847061435900641568717267852309239550206853009021463057851572283500639061743382779907334073926896350263764372737516102187386551242814170610855548491050382678574967152668862227883004100688694285204599343384587231111711912140017478382711012082569512738180968957272901804068838492245715405821721219069121893835580859606977908238643354008308976597052630945957874380249432"),
			},
			want: "this is a test this is a test this is a test this is a test",
		},
		{
			name:    "invalid case, hints are not provided",
			n:       ln.FmpString("145089264118764276482000175726681870278495712"),
			e:       ln.FmpString("65537"),
			ct:      ln.FmpString("12585113701027039647070027075524651900427520559908758474018211051004"),
			max:     4096,
			wantErr: true,
		},
		{
			name:    "invalid case, max not provided",
			n:       ln.FmpString("145089264118764276482000175726681870278495712"),
			e:       ln.FmpString("65537"),
			ct:      ln.FmpString("12585113701027039647070027075524651900427520559908758474018211051004"),
			hints:   []*fmp.Fmpz{ln.FmpString("65537"), ln.FmpString("65537")},
			wantErr: true,
		},
	}

	for _, tc := range tt {
		k, _ := keys.NewRSA(keys.PrivateFromPublic(&keys.FMPPublicKey{
			N: tc.n,
			E: tc.e,
		}), nil, nil, "", false)

		k.Hints = tc.hints
		k.BruteMax = tc.max

		if tc.ct != nil {
			k.CipherText = ln.NumberToBytes(tc.ct)
		}
		ch := make(chan error)
		go Attack([]*keys.RSA{k}, ch)
		err := <-ch
		if err != nil && !tc.wantErr {
			t.Errorf("Attack() failed: %s expected no error got error: %v", tc.name, err)
		}

		if string(k.PlainText) != tc.want && !tc.wantErr {
			t.Errorf("Attack() failed: %s got/want mismatch %s/%s", tc.name, string(k.PlainText), tc.want)
		}
	}

}
