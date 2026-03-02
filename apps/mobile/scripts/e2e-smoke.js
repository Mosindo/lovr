/* eslint-disable no-console */
const API_BASE_URL = process.env.MOBILE_E2E_API_URL || process.env.EXPO_PUBLIC_API_URL || "http://localhost:18080";
const PASSWORD = "Password123";

function uniqueEmail(prefix) {
  return `${prefix}_${Date.now()}_${Math.floor(Math.random() * 100000)}@lovr.test`;
}

async function request(path, init = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
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
  assert(res.status === 201, `register failed for ${email}: ${res.status} ${JSON.stringify(res.payload)}`);
  return res.payload;
}

async function main() {
  console.log(`[e2e-smoke] API: ${API_BASE_URL}`);

  const email1 = uniqueEmail("mobile_e2e_a");
  const email2 = uniqueEmail("mobile_e2e_b");

  const auth1 = await register(email1);
  const auth2 = await register(email2);
  assert(auth1?.token && auth2?.token, "expected tokens after register");

  const login1 = await request("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email: email1, password: PASSWORD })
  });
  assert(login1.status === 200, `login failed: ${login1.status} ${JSON.stringify(login1.payload)}`);
  assert(login1.payload?.user?.id === auth1.user.id, "login user mismatch");

  const discover1 = await request("/discover", {
    method: "GET",
    headers: { Authorization: `Bearer ${auth1.token}` }
  });
  assert(discover1.status === 200, `discover failed: ${discover1.status}`);
  const seesUser2 = Array.isArray(discover1.payload?.users) && discover1.payload.users.some((u) => u.id === auth2.user.id);
  assert(seesUser2, "discover should include second user before interactions");

  const like12 = await request("/likes", {
    method: "POST",
    headers: { Authorization: `Bearer ${auth1.token}` },
    body: JSON.stringify({ toUserId: auth2.user.id })
  });
  assert(like12.status === 200, `like 1->2 failed: ${like12.status}`);
  assert(like12.payload?.matched === false, "first like should not be matched");

  const like21 = await request("/likes", {
    method: "POST",
    headers: { Authorization: `Bearer ${auth2.token}` },
    body: JSON.stringify({ toUserId: auth1.user.id })
  });
  assert(like21.status === 200, `like 2->1 failed: ${like21.status}`);
  assert(like21.payload?.matched === true, "second like should produce match");

  const send = await request(`/chats/${auth2.user.id}/messages`, {
    method: "POST",
    headers: { Authorization: `Bearer ${auth1.token}` },
    body: JSON.stringify({ content: "hello from mobile e2e smoke" })
  });
  assert(send.status === 201, `send message failed: ${send.status}`);

  const listMessages = await request(`/chats/${auth2.user.id}/messages`, {
    method: "GET",
    headers: { Authorization: `Bearer ${auth1.token}` }
  });
  assert(listMessages.status === 200, `list messages failed: ${listMessages.status}`);
  assert(Array.isArray(listMessages.payload?.messages) && listMessages.payload.messages.length > 0, "messages should not be empty");

  const block = await request("/block", {
    method: "POST",
    headers: { Authorization: `Bearer ${auth1.token}` },
    body: JSON.stringify({ toUserId: auth2.user.id })
  });
  assert(block.status === 200, `block failed: ${block.status}`);
  assert(block.payload?.blocked === true, "block response should be blocked=true");

  const likeAfterBlock = await request("/likes", {
    method: "POST",
    headers: { Authorization: `Bearer ${auth2.token}` },
    body: JSON.stringify({ toUserId: auth1.user.id })
  });
  assert(likeAfterBlock.status === 403, `like after block should be 403, got ${likeAfterBlock.status}`);

  const discoverAfter = await request("/discover", {
    method: "GET",
    headers: { Authorization: `Bearer ${auth1.token}` }
  });
  assert(discoverAfter.status === 200, `discover after block failed: ${discoverAfter.status}`);
  const stillSeesUser2 =
    Array.isArray(discoverAfter.payload?.users) && discoverAfter.payload.users.some((u) => u.id === auth2.user.id);
  assert(!stillSeesUser2, "blocked user should no longer appear in discover");

  console.log("[e2e-smoke] PASS register/login/discover/like-match/chat/block");
}

main().catch((err) => {
  console.error(`[e2e-smoke] FAIL ${err instanceof Error ? err.message : String(err)}`);
  process.exit(1);
});
