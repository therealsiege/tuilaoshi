import { createRemoteJWKSet, type JWTVerifyGetKey } from "jose";
import { HttpError } from "./http";
import { PgStore } from "./pg-store";
import { MemoryStore, type Store } from "./store";

// Shared dependencies, created on first use; tests swap them in.

let store: Store | undefined;
let appleKeys: JWTVerifyGetKey | undefined;

export const APPLE_JWKS_URL = "https://appleid.apple.com/auth/keys";

export function getStore(): Store {
  if (store) return store;
  const url = process.env.DATABASE_URL?.trim();
  if (url) {
    store = PgStore.fromUrl(url);
  } else if (process.env.NODE_ENV === "production") {
    throw new HttpError(500, "server misconfigured: DATABASE_URL is not set");
  } else {
    console.warn("DATABASE_URL is not set: using an in-memory store");
    store = new MemoryStore();
  }
  return store;
}

export function getAppleKeys(): JWTVerifyGetKey {
  return (appleKeys ??= createRemoteJWKSet(new URL(APPLE_JWKS_URL)));
}

export function setDepsForTests(d: { store?: Store; appleKeys?: JWTVerifyGetKey }): void {
  store = d.store;
  appleKeys = d.appleKeys;
}
