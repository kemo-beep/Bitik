import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 25,
  duration: "2m",
};

const base = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
  const res = http.get(`${base}/api/v1/public/search?q=shoe&page=1&limit=20`);
  check(res, { "search status 200": (r) => r.status === 200 });
  sleep(1);
}
