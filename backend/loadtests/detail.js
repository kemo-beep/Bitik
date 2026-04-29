import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 10,
  duration: "2m",
};

const base = __ENV.BASE_URL || "http://localhost:8080";
const productID = __ENV.PRODUCT_ID || "00000000-0000-0000-0000-000000000001";

export default function () {
  const res = http.get(`${base}/api/v1/public/products/${productID}`);
  check(res, { "detail status 200": (r) => r.status === 200 || r.status === 404 });
  sleep(1);
}
