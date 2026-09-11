import axios from "axios";

export function readAxiosError(error: unknown, fallback: string) {
    if (axios.isAxiosError<{ error?: { message?: string }; msg?: string; code?: number }>(error)) {
        const status = error.response?.status;
        if (status === 401 || status === 403) {
            return `上游接口鉴权失败（${status}），请检查 API Key 是否正确、是否已过期，以及渠道或模型权限`;
        }
        const responseData = error.response?.data;
        return responseData?.msg || responseData?.error?.message || (status ? `${fallback}：${status}` : fallback);
    }
    return error instanceof Error ? error.message : fallback;
}
