export type ModelChannelProtocol = "openai" | "gemini" | "grok2api" | "doingai" | "metaso" | "apimart" | "kie" | "mimo" | "88api";

export const modelChannelDefaultBaseUrls: Record<ModelChannelProtocol, string> = {
    openai: "https://api.openai.com",
    gemini: "https://generativelanguage.googleapis.com",
    grok2api: "",
    doingai: "https://ai.doingfb.com/v1",
    metaso: "https://metaso.cn/api/minimax",
    apimart: "https://api.apimart.ai/v1",
    kie: "https://api.kie.ai/api/v1",
    mimo: "https://api.xiaomimimo.com",
    "88api": "https://88api.ai/v1",
};

export const modelChannelApiKeyUrls: Partial<Record<ModelChannelProtocol, string>> = {
    doingai: "https://ai.doingfb.com/keys",
};
