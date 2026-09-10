import assert from "node:assert/strict";
import test from "node:test";

import { customChannelProtocolOptions, getChannelDefaultBaseUrl, getCustomChannelProtocolOptions } from "./custom-channel-policy";

test("custom branch channel selector keeps the local four-option order", () => {
    assert.deepEqual(customChannelProtocolOptions, [
        { label: "DoingAI", value: "doingai" },
        { label: "OpenAI", value: "openai" },
        { label: "Gemini", value: "gemini" },
        { label: "自定义", value: "custom" },
    ]);
});

test("legacy protocol remains visible only while editing an existing channel", () => {
    assert.equal(getCustomChannelProtocolOptions().length, 4);
    assert.equal(getCustomChannelProtocolOptions("metaso").at(-1)?.value, "metaso");
    assert.equal(getCustomChannelProtocolOptions("metaso").at(-1)?.label, "历史协议：metaso");
    assert.equal(getCustomChannelProtocolOptions("gemini").length, 4);
});

test("custom protocol leaves the base URL empty", () => {
    assert.equal(getChannelDefaultBaseUrl("custom"), "");
    assert.equal(getChannelDefaultBaseUrl("doingai"), "https://ai.doingfb.com/v1");
});
