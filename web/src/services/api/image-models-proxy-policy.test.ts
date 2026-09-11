import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import test from "node:test";

test("fetchImageModels keeps local model list fetching behind the backend proxy", () => {
    const source = readFileSync(fileURLToPath(new URL("./image.ts", import.meta.url)), "utf8");
    const body = source.match(/export async function fetchImageModels\(config: AiConfig\) \{([\s\S]*?)\n\}/)?.[1] || "";

    assert.match(body, /apiPost<string\[\]>\("\/api\/v1\/models"/);
    assert.doesNotMatch(body, /axios\.get|fetchGeminiModels|fetchAutoDLWorkflows|buildApiUrl/);
});
