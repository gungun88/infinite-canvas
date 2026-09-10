import { modelChannelApiKeyUrls, modelChannelDefaultBaseUrls } from "@/lib/model-channel";

// Keep this local policy separate from the upstream protocol registry.
const primaryChannelOptions = [
    { label: "DoingAI", value: "doingai" },
    { label: "OpenAI", value: "openai" },
    { label: "Gemini", value: "gemini" },
    { label: "自定义", value: "custom" },
] as const;

const visibleChannelOptions = () => primaryChannelOptions.map((option) => ({ ...option }));

export const customChannelProtocolOptions = visibleChannelOptions();

export function getCustomChannelProtocolOptions(currentProtocol?: string) {
    const current = currentProtocol?.trim();
    if (!current || primaryChannelOptions.some((option) => option.value === current.toLowerCase())) return visibleChannelOptions();
    return [...primaryChannelOptions, { label: `历史协议：${current}`, value: current }];
}

export function getChannelDefaultBaseUrl(protocol?: string) {
    const normalized = protocol?.trim().toLowerCase();
    if (!normalized || normalized === "custom") return "";
    return modelChannelDefaultBaseUrls[normalized as keyof typeof modelChannelDefaultBaseUrls] || "";
}

export function getChannelApiKeyUrl(protocol?: string) {
    const normalized = protocol?.trim().toLowerCase();
    return normalized ? modelChannelApiKeyUrls[normalized as keyof typeof modelChannelApiKeyUrls] : undefined;
}
