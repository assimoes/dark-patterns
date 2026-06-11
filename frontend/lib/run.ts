// One adjudication run mock

export type PanelModel = { id: string; name: string; short: string };

export const panelModels: PanelModel[] = [
    { id: "qwen", name: "Qwen3 Next 80B", short: "Qwen" },
    { id: "llama", name: "Llama 3.3 70B", short: "Llama" },
    { id: "deepseek", name: "DeepSeek V4 Flash", short: "DeepSeek" },
    { id: "gptoss", name: "OpenAI OSS 120B", short: "GPT-OSS" },
];

export type PresentDetection = { code: string; modelId: string; evidence: string };

export type AdjReview = {
    id: number;
    gameId: string;
    votedUp: boolean;
    language: string;
    body: string;
    present: PresentDetection[];
};

// det(code, evidence, ...models) — every listed model voted present and cited `evidence`.
const det = (code: string, evidence: string, ...modelIds: string[]): PresentDetection[] =>
    modelIds.map((modelId) => ({ code, modelId, evidence }));

const reviews: AdjReview[] = [
    {
        id: 4821,
        gameId: "archeage",
        votedUp: false,
        language: "english",
        body:
            "Honestly the game is fun at first but it turns into a textbook gacha trap. The drop rate for the SSR character is 0.5%, and the pity timer only kicks in at 200 pulls — that is basically 400 euros to guarantee a single unit. Every single day it nags you to log in or you lose your streak and the weekly bonus. The shop keeps screaming 'LIMITED TIME, 3 hours left!' for a bundle that came back two weeks later anyway. Free players get matched against paying whales and just lose. I had already sunk over 200 euros before I realised what was happening and quitting felt like throwing all of it away.",
        present: [
            ...det(
                "PM-3",
                "The drop rate for the SSR character is 0.5%, and the pity timer only kicks in at 200 pulls",
                "qwen",
                "llama",
                "gptoss",
            ),
            ...det(
                "TM-2",
                "Every single day it nags you to log in or you lose your streak and the weekly bonus",
                "qwen",
                "deepseek",
            ),
            ...det(
                "PM-4",
                "The shop keeps screaming 'LIMITED TIME, 3 hours left!' for a bundle that came back two weeks later",
                "qwen",
                "llama",
                "deepseek",
                "gptoss",
            ),
            ...det("SE-3", "Free players get matched against paying whales and just lose", "llama"),
            ...det(
                "PE-1",
                "I had already sunk over 200 euros before I realised what was happening and quitting felt like throwing all of it away",
                "qwen",
                "gptoss",
            ),
        ],
    },
    {
        id: 4822,
        gameId: "cod",
        votedUp: true,
        language: "english",
        body:
            "Single player campaign was actually great, better than people give it credit for. No pushy microtransactions in the campaign at all — you buy the game and you get the game. The only real annoyance is that the multiplayer supply drops are basically loot boxes you can buy with real money, and the best guns hide behind them. Other than that, fair and complete.",
        present: [
            ...det(
                "PM-3",
                "the multiplayer supply drops are basically loot boxes you can buy with real money",
                "qwen",
                "llama",
                "deepseek",
            ),
            ...det("PM-1", "the best guns hide behind them", "llama"),
        ],
    },
    {
        id: 4823,
        gameId: "poe",
        votedUp: false,
        language: "english",
        body:
            "Used to love this but the monetisation got gross. There is a battle pass every single season and if you skip one season you fall permanently behind, so you feel forced to keep paying 10 euros a month forever. The 'claim daily reward' button is placed exactly where a 'confirm purchase' popup appears a second later, and I bought a 500 gem pack by accident twice. Event items are flagged 'NEVER coming back' to panic you into buying. The ads to double your rewards are basically mandatory if you don't want to grind five times longer.",
        present: [
            ...det(
                "PM-5",
                "There is a battle pass every single season and if you skip one season you fall permanently behind",
                "qwen",
                "llama",
                "deepseek",
                "gptoss",
            ),
            ...det(
                "PM-6",
                "The 'claim daily reward' button is placed exactly where a 'confirm purchase' popup appears a second later, and I bought a 500 gem pack by accident twice",
                "qwen",
                "deepseek",
                "gptoss",
            ),
            ...det(
                "PE-1",
                "Event items are flagged 'NEVER coming back' to panic you into buying",
                "qwen",
                "llama",
                "gptoss",
            ),
            ...det(
                "TM-4",
                "The ads to double your rewards are basically mandatory if you don't want to grind five times longer",
                "llama",
                "deepseek",
            ),
        ],
    },
    {
        id: 4824,
        gameId: "wot",
        votedUp: true,
        language: "english",
        body:
            "It's fine. Grindy like every MMO but nothing felt predatory to me, the cosmetics are optional and clearly priced. Took a while to level but that is just the genre.",
        present: [
            ...det("TM-1", "Grindy like every MMO", "qwen", "llama", "gptoss"),
        ],
    },
];

export const adjRun = {
    id: 2,
    type: "llm_panel" as const,
    taxonomyVersion: 2,
    models: panelModels,
    reviews,
};