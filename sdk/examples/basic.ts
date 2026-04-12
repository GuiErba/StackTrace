import { StackTrace } from "../src";

const st = new StackTrace({
  apiKey: "e1916124-21cd-4603-884b-dba83f61871e",
  service: "example-app",
  baseUrl: "http://localhost:8080",
  environment: "development",
  debug: true,
  batchSize: 5,
  flushIntervalMs: 3000,
});

console.log("Sending logs...\n");

st.info("Application started successfully");
st.info("User logged in", { metadata: { userId: 42, email: "user@test.com" } });
st.warn("Slow database query detected", { metadata: { queryTimeMs: 1200, table: "orders" } });
st.error("Payment gateway timeout", { traceId: "tx-abc-123", metadata: { orderId: 789  } });
st.error("Failed to send notification email", { metadata: { provider: "resend" } });

console.log("Waiting for flush...\n");

setTimeout(async () => {
  await st.shutdown();
  console.log("\nDone! Check your logs at: GET http://localhost:8080/logs");
}, 5000);
