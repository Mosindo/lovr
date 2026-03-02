/* eslint-disable no-console */
const fs = require("fs");
const path = require("path");

const API_BASE_URL = process.env.MOBILE_E2E_API_URL || process.env.EXPO_PUBLIC_API_URL || "http://localhost:18080";
const PASSWORD = process.env.MOBILE_E2E_PASSWORD || "Password123";

function uniqueEmail(prefix) {
  return `${prefix}_${Date.now()}_${Math.floor(Math.random() * 100000)}@lovr.test`;
}

async function request(pathname, init = {}) {
  const response = await fetch(`${API_BASE_URL}${pathname}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init.headers || {})
    }
  });

  let payload = null;
  try {
    payload = await response.json();
  } catch {
    payload = null;
  }

  return { status: response.status, ok: response.ok, payload };
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

async function register(email) {
  const res = await request("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password: PASSWORD })
  });
  assert(res.status === 201, `register failed for ${email}: ${res.status}`);
  return res.payload;
}

async function main() {
  console.log(`[e2e-ui-setup] API: ${API_BASE_URL}`);

  const primaryEmail = uniqueEmail("mobile_ui_primary");
  const matchEmail = uniqueEmail("mobile_ui_match");

  const primary = await register(primaryEmail);
  const match = await register(matchEmail);

  const like1 = await request("/likes", {
    method: "POST",
    headers: { Authorization: `Bearer ${primary.token}` },
    body: JSON.stringify({ toUserId: match.user.id })
  });
  assert(like1.status === 200, `primary like failed: ${like1.status}`);

  const like2 = await request("/likes", {
    method: "POST",
    headers: { Authorization: `Bearer ${match.token}` },
    body: JSON.stringify({ toUserId: primary.user.id })
  });
  assert(like2.status === 200, `match like failed: ${like2.status}`);
  assert(like2.payload?.matched === true, "expected reciprocal like to create match");

  const outDir = path.join(process.cwd(), ".e2e");
  if (!fs.existsSync(outDir)) {
    fs.mkdirSync(outDir, { recursive: true });
  }

  const envJsonPath = path.join(outDir, "maestro-env.json");
  const envPs1Path = path.join(outDir, "maestro-env.ps1");
  const envData = {
    TEST_EMAIL: primaryEmail,
    TEST_PASSWORD: PASSWORD,
    MATCH_EMAIL: matchEmail,
    API_BASE_URL
  };
  fs.writeFileSync(envJsonPath, JSON.stringify(envData, null, 2));
  fs.writeFileSync(
    envPs1Path,
    `$env:TEST_EMAIL='${primaryEmail}'\n$env:TEST_PASSWORD='${PASSWORD}'\n$env:MATCH_EMAIL='${matchEmail}'\n`
  );

  console.log(`[e2e-ui-setup] PASS`);
  console.log(`[e2e-ui-setup] TEST_EMAIL=${primaryEmail}`);
  console.log(`[e2e-ui-setup] MATCH_EMAIL=${matchEmail}`);
  console.log(`[e2e-ui-setup] env json: ${envJsonPath}`);
  console.log(`[e2e-ui-setup] env ps1: ${envPs1Path}`);
}

main().catch((err) => {
  console.error(`[e2e-ui-setup] FAIL ${err instanceof Error ? err.message : String(err)}`);
  process.exit(1);
});
