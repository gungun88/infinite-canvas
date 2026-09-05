import { useUserStore } from "@/stores/use-user-store";

export function requireAiLogin(channelMode: "remote" | "local") {
    const token = useUserStore.getState().token;
    if (token) return token;
    throw new Error(channelMode === "local" ? "请先登录后再使用本地渠道" : "请先登录后再使用云端渠道");
}
