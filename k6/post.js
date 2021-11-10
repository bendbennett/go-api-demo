import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 50,
            timeUnit: '1s',
            duration: '5m',
            preAllocatedVUs: 20,
            maxVUs: 100,
        },
    },
};

export default function () {
    let data = { first_name: "john", last_name: "smith" }

    let res = http.post(`http://${__ENV.HOST}:3000/user`, JSON.stringify(data));
    check(res, { 'status was 200': (r) => r.status == 201 });
}
