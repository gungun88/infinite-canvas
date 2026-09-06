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

export function nextChannelName(channels: Array<{ name?: string | null }>) {
    const maxIndex = channels.reduce((max, channel) => {
        const match = (channel.name || "").trim().match(/^渠道(\d+)$/);
        const index = match ? Number(match[1]) : 0;
        return index > max ? index : max;
    }, 0);
    return `渠道${maxIndex ? maxIndex + 1 : channels.length + 1}`;
}
