import http from 'k6/http';
import { check } from 'k6';

export default function () {
  const url = 'http://airport-lookup-example.com/';
  const payload = JSON.stringify({
    "local_code": "PHX"
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
