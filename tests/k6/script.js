import http from 'k6/http';
import { check } from 'k6';

const baseURL = (__ENV.BASE_URL || 'http://localhost').replace(/\/$/, '');
const airportCode = __ENV.AIRPORT_CODE || 'PHX';
const params = {
  headers: {
    'Content-Type': 'application/json',
  },
};

function lookup() {
  return http.post(
    baseURL + '/',
    JSON.stringify({ local_code: airportCode }),
    params,
  );
}

export function setup() {
  const response = lookup();
  check(response, {
    'warmup status is 200': (r) => r.status === 200,
    'warmup returns requested airport': (r) =>
      r.body.includes('"local_code":"' + airportCode + '"'),
  });
}

export default function () {
  const response = lookup();
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response is successful': (r) => r.body.includes('"ok":true'),
    'response is a cache hit': (r) => r.body.includes('"source":"cache"'),
    'response returns requested airport': (r) =>
      r.body.includes('"local_code":"' + airportCode + '"'),
  });
}
