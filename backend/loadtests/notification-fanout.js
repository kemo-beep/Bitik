import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 10,
  duration: "2m",
};

const base = __ENV.BASE_URL || "http://localhost:8080";
const token = __ENV.INTERNAL_TOKEN || "";

export default function () {
  const payload = {
    dedupe_key: `k6-${__VU}-${__ITER}`,
    payload: { user_id: "00000000-0000-0000-0000-000000000001", event: "order_created" },
  };
  const res = http.post(`${base}/api/v1/internal/jobs/notification-fanout`, JSON.stringify(payload), {
    headers: { "Content-Type": "application/json", "X-Internal-Token": token },
  });
  check(res, { "fanout status": (r) => [200, 202, 401, 403].includes(r.status) });
  sleep(1);
}
