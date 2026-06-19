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
    id: string;
    gameId: string;
    votedUp: boolean;
    language: string;
    body: string;
    present: PresentDetection[];
    modality?: string;
    imageUri?: string;
    description?: string;
};
