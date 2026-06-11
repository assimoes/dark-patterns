export type Family = "TM" | "PM" | "SE" | "PE" | "DR";

export type Pattern = {
    code: string;
    name: string;
    family: Family;
    definition: string;
    examples: string[];
    counterExamples: string[];
};

export const families: Record<Family, string> = {
    TM: "Temporal Manipulation",
    PM: "Predatory Monetization",
    SE: "Social Exploitation",
    PE: "Psychological Exploitation",
    DR: "Deceptive Representation",
};

export const codebook: Pattern[] = [
    {
        code: "TM-1",
        name: "Artificial Extension",
        family: "TM",
        definition:
            "Mechanics that inflate the time required to progress or complete content far beyond what the core gameplay warrants, to boost engagement metrics or drive monetization.",
        examples: [
            "I had to kill the same five enemies for two hours to get one upgrade — pure busywork to inflate playtime.",
            "Levelling past 60 is a slog; everything past that point gives you 1% of the XP you actually need.",
        ],
        counterExamples: [
            "The endgame is challenging and requires patience, but I never felt my time was being padded artificially.",
        ],
    },
    {
        code: "TM-2",
        name: "Scheduled Coercion",
        family: "TM",
        definition:
            "Mechanics that require or pressure the player to play at specific times or intervals dictated by the game, rather than at the player's convenience.",
        examples: [
            "Miss a daily login and your streak resets — you lose the weekly bonus and three days of progress.",
            "Raids only run between 8pm and 10pm, so if you have a real life you're locked out of the best loot.",
        ],
        counterExamples: [
            "There's a weekly tournament every Saturday, but you can join any week — no penalty for skipping.",
        ],
    },
    {
        code: "TM-3",
        name: "Session Manipulation",
        family: "TM",
        definition:
            "Mechanics that artificially extend or constrain individual play sessions, preventing the player from playing at their own pace.",
        examples: [
            "Energy refills at 1 unit every 6 minutes — you literally can't play in a single sitting unless you pay to skip the timer.",
            "There's no pause button mid-mission; if you take a phone call you lose the whole run.",
        ],
        counterExamples: [
            "Online matches obviously can't be paused, but offline mode lets you pause whenever — fair design.",
        ],
    },
    {
        code: "TM-4",
        name: "Forced Ad Exposure",
        family: "TM",
        definition:
            "Requiring or heavily incentivizing the player to watch advertisements, consuming their time in exchange for continued play or in-game rewards.",
        examples: [
            "Every revive shows a 30-second unskippable ad — dying twice in a row means a minute of advertising before I can play again.",
            "You CAN play without ads, but progress is 5x slower unless you watch a video every match. Not really optional.",
        ],
        counterExamples: [
            "There's an 'watch ad for a daily bonus' button I can ignore without consequence — completely optional reward.",
        ],
    },
    {
        code: "PM-1",
        name: "Pay-to-Progress",
        family: "PM",
        definition:
            "Systems where spending real money provides gameplay advantages or unlocks content/progression that non-paying players cannot access or must invest disproportionate time to reach.",
        examples: [
            "Premium players get 2x damage and faster movement — F2P PvP is basically unplayable past silver rank.",
            "You can't even unlock the second city without buying the season pass. Hard paywall after the prologue.",
        ],
        counterExamples: [
            "There are cosmetic skins you can buy but they don't affect gameplay at all — purely visual stuff.",
        ],
    },
    {
        code: "PM-2",
        name: "Currency Obfuscation",
        family: "PM",
        definition:
            "Using virtual or multi-layered currency systems to obscure the real-world cost of in-game purchases, making it harder for players to evaluate how much they are actually spending.",
        examples: [
            "1 'gem' is supposedly 1 cent but bundles sell at 99/249/499 gems — try doing the math on what a skin actually costs.",
            "You buy 'platinum', spend platinum on 'crystals', spend crystals on 'tickets'. I lost track of how much I'd spent.",
        ],
        counterExamples: [
            "Items have clear €/$ prices in the store, no virtual currency middleman — refreshing transparency.",
        ],
    },
    {
        code: "PM-3",
        name: "Gambling Mechanics",
        family: "PM",
        definition:
            "Systems that use randomized rewards purchased with real money, exploiting the same psychological mechanisms as gambling (variable ratio reinforcement, near-misses, etc.).",
        examples: [
            "Drop rate for the SSR character is 0.5%, but the pity timer kicks in at 200 pulls — roughly €400 to guarantee one.",
            "Spinning the wheel costs gems; you get whatever it lands on, and 'try again' is one tap away. Pure slot machine.",
        ],
        counterExamples: [
            "Battle pass rewards are fixed and listed in advance — you know exactly what you get for each tier. No RNG.",
        ],
    },
    {
        code: "PM-4",
        name: "Price Manipulation",
        family: "PM",
        definition:
            "Techniques that distort the player's perception of value or cost through anchoring, artificial scarcity, or depreciation of prior purchases.",
        examples: [
            "The 'Mega Bundle' is shown crossed-out at €99.99 right next to the €19.99 'special offer' — textbook anchoring.",
            "Shop says 'LIMITED — 3 hours left!' but the same skin came back two weeks later in another 'limited' rotation.",
        ],
        counterExamples: [
            "Prices are static and shown in your local currency without any 'sale' framing or countdown timers.",
        ],
    },
    {
        code: "PM-5",
        name: "Recurring / Compounding Charges",
        family: "PM",
        definition:
            "Monetization systems that create ongoing or escalating financial commitments beyond a single purchase.",
        examples: [
            "The battle pass is €10 every season; skip one season and you fall permanently behind the rewards track.",
            "Premium subscription renews monthly and silently upcharged me when the price rose — no email warning.",
        ],
        counterExamples: [
            "There's an optional one-time DLC purchase, no recurring charges or subscriptions of any kind.",
        ],
    },
    {
        code: "PM-6",
        name: "Interface Monetization Traps",
        family: "PM",
        definition:
            "User interface designs that trick, nudge, or facilitate accidental purchases through deceptive placement, confusing flows, or missing confirmation steps.",
        examples: [
            "The 'buy now' button is exactly where the 'close' button was two seconds earlier — I bought a 50-gem pack by accident.",
            "Claiming a daily reward triggers a 'CONFIRM SPEND 200 GEMS' dialog with a misleading green button.",
        ],
        counterExamples: [
            "Purchases require a typed PIN and a second confirmation screen — basically impossible to misclick a real-money buy.",
        ],
    },
    {
        code: "SE-1",
        name: "Recruitment Pressure",
        family: "SE",
        definition:
            "Mechanics that incentivize or require players to recruit new players from their social network in order to progress, access features, or receive rewards.",
        examples: [
            "To unlock the guild boss you need to invite 5 friends who actually start playing — chat is full of people begging for invites.",
            "The game posts to your friends list automatically saying 'Antonio is recruiting!' without asking for permission.",
        ],
        counterExamples: [
            "There's a referral code that gives a small one-time bonus if used, but it's tucked away and clearly optional.",
        ],
    },
    {
        code: "SE-2",
        name: "Social Obligation",
        family: "SE",
        definition:
            "Mechanics that create feelings of guilt, duty, or reciprocal debt toward other players, making the player feel they must play or spend to avoid letting others down.",
        examples: [
            "Your guild loses points if you don't log in daily, so the other 49 members guilt you into playing.",
            "If your co-op partner doesn't show up you BOTH lose three days of progress — pure peer pressure to keep logging in.",
        ],
        counterExamples: [
            "Co-op missions reward both players when you play together, but you're not penalised at all for soloing.",
        ],
    },
    {
        code: "SE-3",
        name: "Competitive Pressure",
        family: "SE",
        definition:
            "Systems that exploit players' competitive nature by creating environments where spending money is the primary way to compete, or by using matchmaking to showcase the advantages of paying.",
        examples: [
            "Free players keep getting matched against premium subscribers and losing — the message is clear: pay or lose.",
            "The leaderboard is locked behind a €30 'competitive pack' that gives a 15% stat boost over anyone who didn't buy it.",
        ],
        counterExamples: [
            "Ranked PvP uses MMR-based matchmaking; paying doesn't influence skill brackets at all.",
        ],
    },
    {
        code: "PE-1",
        name: "Loss Aversion Exploitation",
        family: "PE",
        definition:
            "Mechanics that leverage the player's fear of losing progress, items, or opportunities — exploiting the principle that losses feel more painful than equivalent gains.",
        examples: [
            "I've spent €200 over six months — quitting now feels like throwing all that money away, which is exactly what they want.",
            "Limited-time event ends in 6 hours, items will NEVER return — anxiety-driven purchase, classic FOMO.",
        ],
        counterExamples: [
            "All cosmetics rotate back into the shop every couple of months, so missing one event isn't a big deal.",
        ],
    },
    {
        code: "PE-2",
        name: "Completionism Exploitation",
        family: "PE",
        definition:
            "Mechanics that exploit the player's desire to collect, complete, or achieve everything, creating artificial compulsion to engage beyond enjoyment.",
        examples: [
            "The 100% achievement requires collecting 500 of a currency that drops at 1% rate — designed for completionists.",
            "The collection tab shows 87/100 done, dangling that last 13% to keep me grinding past when I stopped enjoying it.",
        ],
        counterExamples: [
            "There are optional collectibles but the game makes it clear they're not required and won't affect the story.",
        ],
    },
    {
        code: "PE-3",
        name: "Cognitive Bias Exploitation",
        family: "PE",
        definition:
            "Mechanics that exploit specific cognitive biases to make players misjudge probabilities, costs, or their own agency.",
        examples: [
            "You 'pick' a chest from three identical boxes — but the reward was determined before you tapped. Illusion of control.",
            "Drop rate is shown as '~15%' but in practice feels closer to 3% — they exploit optimism bias by showcasing rare drops.",
        ],
        counterExamples: [
            "Drop rates are published transparently and have been verified by community datamining.",
        ],
    },
    {
        code: "PE-4",
        name: "Sensory Manipulation",
        family: "PE",
        definition:
            "Using visual effects, sound design, animations, or other sensory elements to trigger emotional responses that override rational decision-making about spending or engagement.",
        examples: [
            "Opening a loot box has glittery near-miss animations and a triumphant fanfare even for trash drops — engineered to feel like a win.",
            "Every purchase triggers a 5-second celebration screen with confetti and rising chords, like a slot machine win.",
        ],
        counterExamples: [
            "Purchases just show a quiet 'item added to inventory' notification with no fanfare or animation.",
        ],
    },
    {
        code: "DR-1",
        name: "Misleading Marketing",
        family: "DR",
        definition:
            "Advertising, store listings, or promotional materials that misrepresent the actual gameplay, graphics, content, or experience the player will have.",
        examples: [
            "The Facebook ads show a totally different game — actual gameplay is nothing like the promised tactical battles.",
            "Trailer features mechanics and characters that aren't even in the released game. Pure bait.",
        ],
        counterExamples: [
            "Marketing accurately represents the gameplay; what you see in trailers is what you get.",
        ],
    },
    {
        code: "DR-2",
        name: "Post-Purchase Deception",
        family: "DR",
        definition:
            "Changes or discrepancies between what was promised/previewed at the point of purchase and what the player actually receives or experiences after spending.",
        examples: [
            "I paid €25 for the meta hero and they nerfed it to the ground two weeks later — never refunded the difference.",
            "The cosmetic looks completely different in-game than in the store preview — much lower quality.",
        ],
        counterExamples: [
            "When they balance-changed a character they offered a free refund window — fair treatment of paying players.",
        ],
    },
];

export const patternByCode = (code: string): Pattern =>
    codebook.find((p) => p.code === code) ?? codebook[0];