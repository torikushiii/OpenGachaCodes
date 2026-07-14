package mobalytics

import (
	"testing"
	"time"
)

func TestParseMarkdownActiveCodes(t *testing.T) {
	page := `## Endfield Active Promo Codes

| ### Latest Codes | ### Rewards |
| --- | --- |
| ENDFIELDRENEW **1.4 Special Program Code** | * **Sticker: Safety Helmet** * **71600 T-Credits** **VALID UNTIL THE END OF THE VERSION 1.4!** |
| ZAU2SYXHWX5L4ZH | * 5000 T-Credits * 10 Arms INSP Kit |
| RETURNOFALL **Expired** | * **Oroberyl x500** |
| ENDFIELDGIFT | * **T-Creds x13,000** * **Advanced Combat Record x2** |
| ENDFIELD4PC **PC Only** | * **Oroberyl x150** |

## How to Redeem`
	candidates, err := parseMarkdown(page, "test", time.Now())
	if err != nil || len(candidates) != 4 {
		t.Fatalf("candidates=%+v err=%v", candidates, err)
	}
	if candidates[0].Code != "ENDFIELDRENEW" || candidates[0].Rewards[1] != "T-Creds ×71,600" {
		t.Fatalf("candidate=%+v", candidates[0])
	}
	if candidates[1].Rewards[0] != "T-Creds ×5,000" || candidates[1].Rewards[1] != "Arms INSP Kit ×10" {
		t.Fatalf("candidate=%+v", candidates[1])
	}
	if candidates[2].Code != "ENDFIELDGIFT" || candidates[2].Rewards[0] != "T-Creds ×13,000" {
		t.Fatalf("candidate=%+v", candidates[2])
	}
	if candidates[3].Code != "ENDFIELD4PC" || candidates[3].Rewards[0] != "Oroberyl ×150" {
		t.Fatalf("candidate=%+v", candidates[3])
	}
}
