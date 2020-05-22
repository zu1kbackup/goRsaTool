package commonfactor

import (
	"testing"

	"github.com/sourcekris/goRsaTool/keys"
	"github.com/sourcekris/goRsaTool/ln"
	"github.com/sourcekris/goRsaTool/utils"

	fmp "github.com/sourcekris/goflint"
)

func TestAttack(t *testing.T) {
	tt := []struct {
		name string
		n1   *fmp.Fmpz
		n2   *fmp.Fmpz
		want *fmp.Fmpz
	}{
		{
			name: "vulnerable key pair expected to factor",
			n1:   ln.FmpString("19921235092885718941460705440825384766889688808288887410363133380298855225855169425287727692673393934178475897827294754849374197720137987960688344110965591170210781289465934066636174381897662365287852177655361788277493432034948523391098040047343547730711993597167763378414064146096938364866043496305522399087408043989884908020018692642580328124229280044641587303382021351178359247138833154554633679728011082348580310030397185519399752172648533232524073066593512844762640362921013193085942163836240699748575123895338983597279867874621482246403835899398515327632824816267688090966829191631224063682485914382314998195093"),
			n2:   ln.FmpString("22281454606178185475137713421838422701543711268688600199661211611180627857676287178299712404685904372784253912486518309166107347902668817333387309917713878185701525779283063877318406271407207356695157218976821377797726991423192800200038862274192839464396744870595855658571673885678865944463809042500492800193755481497663544377666279577049151233765472181498228853733312890990468820942647689943230580776756954044828448094549187428360616039917736728741158185566675010288835722749075283482869482557110351806822719324373000017117153101570619871972625144670079798850809870562279085243502354929201076164300122928273223973813"),
			want: ln.FmpString("146566651445893368688905763456764452337838032763682676221025945682991649793340026890854472049371592346730454191221850371408406581475418579008881111571092173530748331667107582622861309727150160914480781841205155449584530166428770678446245420268299373990760393892275516496045323891286171163252445865368303271017"),
		},
	}

	for _, tc := range tt {
		k1, _ := keys.NewRSA(keys.PrivateFromPublic(&keys.FMPPublicKey{
			N: tc.n1,
			E: fmp.NewFmpz(3),
		}), nil, nil, "", false)

		k2, _ := keys.NewRSA(keys.PrivateFromPublic(&keys.FMPPublicKey{
			N: tc.n2,
			E: fmp.NewFmpz(3),
		}), nil, nil, "", false)

		err := Attack([]*keys.RSA{k1, k2})
		if err != nil {
			t.Errorf("Attack() failed: %s expected no error got error: %v", tc.name, err)
		}

		if !utils.FoundP(tc.want, k1.Key.Primes) {
			t.Errorf("Attack() failed: %s expected primes not found - got %v wanted %v", tc.name, k1.Key.Primes, tc.want)
		}
	}
}