import http from 'k6/http';
import { check, sleep } from 'k6';

// Base URLs
// When you run k6 inside a Docker container, 
// localhost refers to the container itself, not your host machine.
// Your Go API is running on your host (outside the container),
// therefore we have to use host.docker.internal to talk to it
const BASE = 'http://host.docker.internal:8080';

// Define test options
export const options = {
  stages: [
    { duration: '1m', target: 4 },
    { duration: '5m', target: 4 },

    { duration: '1m', target: 8 },
    { duration: '5m', target: 8 },

    { duration: '1m', target: 16 },
    { duration: '5m', target: 16 },

    { duration: '1m', target: 32 },
    { duration: '5m', target: 32 },

    { duration: '1m', target: 64 },
    { duration: '5m', target: 64 },
  ],
};

// Helper to generate unique FAA per VU/iteration
function uniqueFAA() {
  const rand = Math.floor(Math.random() * 900) + 100; // 3-digit
  return `A${__VU}${__ITER}${rand}`.slice(0, 10);
}

export default function () {
  let res;

  // --- HEALTH CHECK ---
  res = http.get(`${BASE}/health`);
  check(res, {
    'HEALTH CHECK status is 200': (r) => r.status === 200,
  });
  sleep(1);

  // --- GET ALL AIRPORTS ---
  res = http.get(`${BASE}/airports`);
  check(res, {
    'GET ALL AIRPORTS status is 200': (r) => r.status === 200,
  });
  sleep(1);
  
  // --- GET AIRPORT BY FAA ---
  res = http.get(`${BASE}/airport/16A`);
  check(res, {
    'GET AIRPORT BY FAA status is 200': (r) => r.status === 200,
  });
  sleep(1);

  // --- SYNC AIRPORT BY FAA ---
  res = http.post(`${BASE}/sync/16A`, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, {
    'SYNC AIRPORT BY FAA status is 200': (r) => r.status === 200,
  });
  sleep(1);

  // We don't have to check the sync all since it shouldn't be done actively

  // --- CREATE airport with unique FAA ---
  const faa = uniqueFAA();
  const createPayload = JSON.stringify({
    site_number: "XYZ",
    facility_name: "PT. XYZ",
    faa_ident: faa,
    icao_ident: "ABC",
    state: "AK",
    state_full: "ALASKA",
    county: "BETHEL",
    city: "JAKARTA",
    ownership: "PU",
    use: "PU",
    manager: "VYN CEZELIANO",
    manager_phone: "(666) 021-021021",
    latitude: "60-54-21.6000N",
    longitude: "162-26-26.1000W",
    status: "O",
    weather: ""
  });

  res = http.post(`${BASE}/airport`, createPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, {
    'CREATE AIRPORT status is 200': (r) => r.status === 200,
  });
  sleep(1);

  // --- UPDATE airport with same FAA ---
  const updatePayload = JSON.stringify({
    site_number: "XYZ",
    facility_name: "UPDATED FACILITY",
    faa_ident: faa,
    icao_ident: "PPIT",
    state: "AK",
    state_full: "ALASKA",
    county: "BETHEL",
    city: "NUNAPITCHUK",
    ownership: "PU",
    use: "PU",
    manager: "JOSEPH LARAUX",
    manager_phone: "(907) 543-2498",
    latitude: "60-54-21.6000N",
    longitude: "162-26-26.1000W",
    status: "O",
    weather: "Light snow"
  });

  res = http.put(`${BASE}/airport`, updatePayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, {
    'UPDATE AIRPORT status is 200': (r) => r.status === 200,
  });
  sleep(1);

  // --- DELETE airport with same FAA ---
  res = http.del(`${BASE}/airport/${faa}`);
  check(res, {
    'DELETE AIRPORT status is 200': (r) => r.status === 200,
  });
  sleep(1);
}
