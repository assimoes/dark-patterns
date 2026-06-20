package research

import (
	"strings"
	"testing"
)

func sampleProfile() Profile {
	full := "free-to-play with cosmetic microtransactions"
	anticipated := "anticipated battle pass, unconfirmed"
	return Profile{
		Game:      "Warframe",
		Developer: "Digital Extremes",
		Genre:     "Third-person looter shooter (online-only)",
		CoreLoop:  "run missions, collect resources, craft and level gear",
		BusinessModel: BusinessModel{
			CurrentState:           "free-to-play",
			AtFullRelease:          &full,
			RealMoneyScope:         "platinum buys cosmetics, slots, and market shortcuts",
			AnticipatedUnconfirmed: &anticipated,
		},
		Resources: []Resource{
			{Name: "Credits", Acquisition: "earned-through-play", RealMoneyPurchasable: false, Purpose: "in-game economy currency for crafting and trading"},
			{Name: "Platinum", Acquisition: "both", RealMoneyPurchasable: true, Purpose: "premium store currency for cosmetics and slots"},
		},
		NotableMechanics: []string{"Trading between players", "Time-gated crafting"},
		Sources:          []string{"https://example.com/warframe"},
	}
}

func TestRenderDeterministicAndSeparatesCurrencies(t *testing.T) {
	p := sampleProfile()
	a := Render(p)
	b := Render(p)
	if a != b {
		t.Fatal("render is not deterministic")
	}
	for _, want := range []string{
		"earned through play",
		"can be bought with real money",
		"not bought with real money",
		"in-game economy currency",
		"premium store currency",
	} {
		if !strings.Contains(a, want) {
			t.Errorf("rendered text missing %q\n---\n%s", want, a)
		}
	}
	// the speculative field must never reach the panel's context.
	if strings.Contains(a, "anticipated") {
		t.Errorf("rendered text should omit anticipated/unconfirmed, got\n%s", a)
	}
}

func TestScanValence(t *testing.T) {
	if flags := ScanValence(Render(sampleProfile())); len(flags) != 0 {
		t.Errorf("neutral description should have no valence flags, got %v", flags)
	}

	flags := ScanValence("the monetization is predatory and the devs are notoriously greedy, but generous with cosmetics")
	got := map[string]bool{}
	for _, f := range flags {
		got[f.Term] = true
	}
	for _, want := range []string{"predatory", "greedy", "generous"} {
		if !got[want] {
			t.Errorf("expected valence flag %q, got %v", want, flags)
		}
	}
}

func TestValidate(t *testing.T) {
	if err := sampleProfile().Validate(); err != nil {
		t.Fatalf("valid profile rejected: %v", err)
	}

	bad := sampleProfile()
	bad.Game = ""
	if err := bad.Validate(); err == nil {
		t.Error("expected missing-game error")
	}

	badAcq := sampleProfile()
	badAcq.Resources[0].Acquisition = "free"
	if err := badAcq.Validate(); err == nil {
		t.Error("expected invalid-acquisition error")
	}
}
