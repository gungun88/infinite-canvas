import assert from "node:assert/strict";
import test from "node:test";
import { AxiosError } from "axios";
import { readAxiosError } from "./errors";

test("readAxiosError hides upstream auth response details", () => {
    const error = new AxiosError("Request failed", "ERR_BAD_REQUEST", undefined, undefined, {
        data: { error: { message: "invalid api key: sk-secret" } },
        status: 401,
        statusText: "Unauthorized",
        headers: {},
        config: {},
    } as never);

    const message = readAxiosError(error, "读取模型失败");

    assert.equal(message, "上游接口鉴权失败（401），请检查 API Key 是否正确、是否已过期，以及渠道或模型权限");
    assert.equal(message.includes("sk-secret"), false);
});
