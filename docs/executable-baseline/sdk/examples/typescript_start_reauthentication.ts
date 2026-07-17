export async function startReauthentication(client: AuthenticationClient) {
  return client.startAuthentication({
    clientId: "clients/example",
    subjectHint: "user@example.test",
    reason: "REFRESH_UNAVAILABLE",
    requestedAal: "AAL2",
    idempotencyKey: crypto.randomUUID(),
  });
}
