package research

import (
	"fmt"
	"strings"
)

// Render turns a profile into the plain text injected into the annotation prompt and hashed into
// config_digest. it reads as a reference glossary, not an assessment: the point is to let the panel
// resolve what the review's wording refers to (e.g. whether a named currency is earned or bought), never
// to feed it a monetisation verdict. it is deterministic (fixed order) so the same profile always yields
// the same text and hash. business_model.anticipated_unconfirmed is intentionally left out: it is
// speculative and for the human reviewer, not the panel.
func Render(p Profile) string {
	var b strings.Builder

	head := p.Game
	var qual []string
	if g := strings.TrimSpace(p.Genre); g != "" {
		qual = append(qual, g)
	}
	if d := strings.TrimSpace(p.Developer); d != "" {
		qual = append(qual, "by "+d)
	}
	if len(qual) > 0 {
		head += " (" + strings.Join(qual, ", ") + ")"
	}
	fmt.Fprintf(&b, "Reference facts for %s.\n", head)

	if strings.TrimSpace(p.CoreLoop) != "" {
		fmt.Fprintf(&b, "What the player does: %s\n", p.CoreLoop)
	}

	bm := p.BusinessModel
	model := bm.CurrentState
	if bm.AtFullRelease != nil && strings.TrimSpace(*bm.AtFullRelease) != "" {
		model += fmt.Sprintf(" (at full release: %s)", *bm.AtFullRelease)
	}
	if strings.TrimSpace(model) != "" {
		fmt.Fprintf(&b, "Monetisation model: %s\n", model)
	}
	if strings.TrimSpace(bm.RealMoneyScope) != "" {
		fmt.Fprintf(&b, "What real money can buy: %s\n", bm.RealMoneyScope)
	}

	if len(p.Resources) > 0 {
		b.WriteString("Currencies and resources — what each term refers to:\n")
		for _, r := range p.Resources {
			fmt.Fprintf(&b, "- %s — %s; %s. %s\n",
				r.Name, acquisitionPhrase(r.Acquisition), realMoneyPhrase(r.RealMoneyPurchasable), r.Purpose)
		}
	}

	if len(p.NotableMechanics) > 0 {
		b.WriteString("Mechanics that may be referenced:\n")
		for _, m := range p.NotableMechanics {
			fmt.Fprintf(&b, "- %s\n", m)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func acquisitionPhrase(a string) string {
	switch a {
	case "earned-through-play":
		return "earned through play"
	case "purchasable":
		return "obtained by purchase"
	case "both":
		return "earned through play or purchased"
	default:
		return a
	}
}

func realMoneyPhrase(purchasable bool) string {
	if purchasable {
		return "can be bought with real money"
	}
	return "not bought with real money"
}
