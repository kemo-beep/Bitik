import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 10,
  duration: "2m",
};

const base = __ENV.BASE_URL || "http://localhost:8080";
const token = __ENV.ACCESS_TOKEN || "";

export default function () {
  const params = { headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } };
  const cartRes = http.get(`${base}/api/v1/buyer/cart`, params);
  check(cartRes, { "cart status": (r) => r.status === 200 || r.status === 401 });

  const sessionRes = http.post(`${base}/api/v1/buyer/checkout/sessions`, JSON.stringify({}), params);
  check(sessionRes, { "checkout session status": (r) => [201, 400, 401].includes(r.status) });
  sleep(1);
}
