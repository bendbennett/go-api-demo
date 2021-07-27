import http from 'k6/http';
import { check } from 'k6';

export let options = {
    stages: [
        { duration: '120s', target: 200 },
        { duration: '60s', target: 200 },
        { duration: '120s', target: 0 },
    ],
};

export default function () {
    let res = http.get(`http://${__ENV.HOST}:3000/user`);
    check(res, { 'status was 200': (r) => r.status == 200 });
}