// Package research runs the one-off game-research call at onboarding: a single OpenRouter request with
// web search that returns a strictly neutral, structured profile of a game. the profile is stored as a
// draft for human approval; the annotation pipeline only ever reads an approved, frozen version.
package research

import (
	"fmt"
	"strings"
)

// Profile is the structured contract the research call must return. resources is the load-bearing field
// for the disambiguation problem: each currency/resource states how it is acquired and whether real money
// buys it, separating in-game economy currencies from premium/store ones.
type Profile struct {
	Game             string        `json:"game"`
	Developer        string        `json:"developer"`
	Genre            string        `json:"genre"`
	CoreLoop         string        `json:"core_loop"`
	BusinessModel    BusinessModel `json:"business_model"`
	Resources        []Resource    `json:"resources"`
	NotableMechanics []string      `json:"notable_mechanics"`
	Sources          []string      `json:"sources"`
}

// BusinessModel describes how the game makes money, plainly. nullable fields use a pointer so "not
// applicable" stays distinct from an empty assertion.
type BusinessModel struct {
	CurrentState           string  `json:"current_state"`
	AtFullRelease          *string `json:"at_full_release"`
	RealMoneyScope         string  `json:"real_money_scope"`
	AnticipatedUnconfirmed *string `json:"anticipated_unconfirmed"`
}

// Resource is one currency or resource: how it is acquired, whether real money buys it, and what it is for.
type Resource struct {
	Name                 string `json:"name"`
	Acquisition          string `json:"acquisition"`
	RealMoneyPurchasable bool   `json:"real_money_purchasable"`
	Purpose              string `json:"purpose"`
}

// the acquisition values a resource may take.
var validAcquisition = map[string]bool{
	"earned-through-play": true,
	"purchasable":         true,
	"both":                true,
}

// Validate rejects a profile that does not meet the contract, so a malformed research result is surfaced
// to the reviewer rather than frozen. it checks the required fields and the resource acquisition enum.
func (p Profile) Validate() error {
	var missing []string
	if strings.TrimSpace(p.Game) == "" {
		missing = append(missing, "game")
	}
	if strings.TrimSpace(p.Genre) == "" {
		missing = append(missing, "genre")
	}
	if strings.TrimSpace(p.CoreLoop) == "" {
		missing = append(missing, "core_loop")
	}
	if strings.TrimSpace(p.BusinessModel.CurrentState) == "" {
		missing = append(missing, "business_model.current_state")
	}
	if strings.TrimSpace(p.BusinessModel.RealMoneyScope) == "" {
		missing = append(missing, "business_model.real_money_scope")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	for i, r := range p.Resources {
		if strings.TrimSpace(r.Name) == "" {
			return fmt.Errorf("resources[%d]: empty name", i)
		}
		if !validAcquisition[r.Acquisition] {
			return fmt.Errorf("resources[%d] %q: acquisition must be earned-through-play, purchasable, or both, got %q", i, r.Name, r.Acquisition)
		}
		if strings.TrimSpace(r.Purpose) == "" {
			return fmt.Errorf("resources[%d] %q: empty purpose", i, r.Name)
		}
	}

	return nil
}

// schemaJSON is the schema the research call is asked to match, embedded verbatim in the user prompt.
const schemaJSON = `{
  "game": "string",
  "developer": "string",
  "genre": "string — factual genre/format, e.g. 'Action RPG (isometric, online-only)'",
  "core_loop": "string — what the player does moment to moment, factual",
  "business_model": {
    "current_state": "string — e.g. paid early access, F2P, subscription, premium",
    "at_full_release": "string or null — if the model differs at full release",
    "real_money_scope": "string — exactly what real money can buy; state plainly whether gameplay power/progression is sold",
    "anticipated_unconfirmed": "string or null — reported-but-unconfirmed monetization features, each marked, for the human reviewer to verify"
  },
  "resources": [
    {
      "name": "string — e.g. 'Currency orbs (Chaos, Exalted, Divine)'",
      "acquisition": "one of: 'earned-through-play' | 'purchasable' | 'both'",
      "real_money_purchasable": true,
      "purpose": "string — what it's used for; if it's an in-game economy currency vs a premium/store currency, say so explicitly"
    }
  ],
  "notable_mechanics": [
    "string — each mechanic stated as existing, not judged (e.g. 'League system: periodic economy resets')"
  ],
  "sources": [
    "string — URLs the description is grounded in"
  ]
}`
