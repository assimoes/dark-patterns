// the structured neutral profile the research call returns and the reviewer edits. resources is the
// load-bearing field: each currency states how it's acquired and whether real money buys it.
export type Acquisition = "earned-through-play" | "purchasable" | "both";

export type Resource = {
    name: string;
    acquisition: Acquisition;
    real_money_purchasable: boolean;
    purpose: string;
};

export type BusinessModel = {
    current_state: string;
    at_full_release: string | null;
    real_money_scope: string;
    anticipated_unconfirmed: string | null;
};

export type DescriptionProfile = {
    game: string;
    developer: string;
    genre: string;
    core_loop: string;
    business_model: BusinessModel;
    resources: Resource[];
    notable_mechanics: string[];
    sources: string[];
};

export type ValenceFlag = { term: string; index: number };

// the latest version's state for a game. drives the games-list badge and the review view.
export type DescriptionStatus =
    | "none"
    | "draft"
    | "approved"
    | "superseded"
    | "invalid"
    | "error";

// one version of a game's description. profile/sources/valenceFlags arrive parsed from the api.
export type GameDescription = {
    id: number;
    externalGameId: string;
    version: number;
    status: DescriptionStatus;
    profile: DescriptionProfile;
    renderedText: string;
    researchModel: string;
    sources: string[];
    valenceFlags: ValenceFlag[];
    error: string;
    createdAt: string;
    approvedAt: string;
};

// an empty profile, for an invalid/error draft or a fresh editor.
export const emptyProfile: DescriptionProfile = {
    game: "",
    developer: "",
    genre: "",
    core_loop: "",
    business_model: {
        current_state: "",
        at_full_release: null,
        real_money_scope: "",
        anticipated_unconfirmed: null,
    },
    resources: [],
    notable_mechanics: [],
    sources: [],
};
