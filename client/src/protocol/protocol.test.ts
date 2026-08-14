import { describe, expect, it } from "vitest";
import { createEnvelope, decodeEnvelope, encodeEnvelope } from "./envelope";
import { validateClientEnvelope, validateServerEnvelopeType } from "./validator";
import { parseServerEnvelope } from "./parser";

describe("protocol envelope", () => {
  it("encodes and decodes an envelope", () => {
    const envelope = createEnvelope("message", { recipient_id: "bob", content: "hi" });
    const json = encodeEnvelope(envelope);
    const decoded = decodeEnvelope(json);
    expect(decoded.version).toBe(1);
    expect(decoded.type).toBe("message");
  });

  it("rejects malformed JSON", () => {
    expect(() => decodeEnvelope("not json")).toThrow();
  });

  it("rejects a non-envelope object", () => {
    expect(() => decodeEnvelope('{"foo":"bar"}')).toThrow();
  });
});

describe("protocol validator", () => {
  it("accepts client-supported types", () => {
    expect(validateClientEnvelope(createEnvelope("message", {}))).toBe(true);
    expect(validateClientEnvelope(createEnvelope("group_message", {}))).toBe(true);
    expect(validateClientEnvelope(createEnvelope("sender_key_distribution", {}))).toBe(true);
    expect(validateClientEnvelope(createEnvelope("profile_picture_updated", {}))).toBe(true);
  });

  it("rejects presence on the client send path", () => {
    expect(validateClientEnvelope(createEnvelope("presence", {}))).toBe(false);
  });

  it("accepts server-originated types", () => {
    for (const type of [
      "message",
      "group_message",
      "sender_key_distribution",
      "profile_picture_updated",
      "presence_snapshot",
      "presence",
      "error",
    ]) {
      expect(validateServerEnvelopeType(type)).toBe(true);
    }
  });
});

describe("protocol parser", () => {
  it("parses a server envelope", () => {
    const env = parseServerEnvelope('{"version":1,"type":"presence_snapshot","payload":{"online":["a"]}}');
    expect(env.type).toBe("presence_snapshot");
  });

  it("rejects malformed payload", () => {
    expect(() => parseServerEnvelope("null")).toThrow();
  });
});
