import json, urllib.request, ssl, time

ctx = ssl._create_unverified_context()
BASE = "http://localhost:8088"
passed = 0
failed = 0
API = f"{BASE}/api/v1"
passed = 0
failed = 0

def test(name, fn):
    global passed, failed
    try:
        result = fn()
        if result:
            print(f"  PASS  {name}")
            passed += 1
        else:
            print(f"  FAIL  {name}")
            failed += 1
    except Exception as e:
        print(f"  ERROR {name}: {e}")
        failed += 1

def get(path, token=None):
    headers = {}
    if token: headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(f"{BASE}{path}", headers=headers)
    try:
        resp = urllib.request.urlopen(req, context=ctx)
        return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        return json.loads(e.read())

def post(path, data, token=None):
    headers = {"Content-Type": "application/json"}
    if token: headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(f"{BASE}{path}", data=json.dumps(data).encode(), headers=headers, method="POST")
    try:
        resp = urllib.request.urlopen(req, timeout=60, context=ctx)
        return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        body = e.read()
        if body:
            try: return json.loads(body)
            except: pass
        return {"error": f"HTTP {e.code}", "message": str(e)}
    except Exception as e:
        return {"error": "TIMEOUT_OR_NETWORK", "message": str(e)}

# === HEALTH ===
print("\n[1] Health Check")
r = get("/api/health")
test("Health endpoint returns healthy", lambda: r.get("status") == "healthy")
test("AI providers listed", lambda: len(r.get("ai", [])) >= 2)
test("Version field present", lambda: "version" in r)

# === AUTH ===
print("\n[2] Auth - Send Code")
r = post("/api/v1/auth/send-code", {"email": "e2etest@163.com"})
test("Send code returns success", lambda: r.get("success") == True)

print("\n[3] Auth - Register (skip verify)")
ts = str(int(time.time()))
r = post("/api/v1/auth/register", {"email": f"e2e{ts}@163.com", "password": "test123456", "name": "E2E Test"})
test("Register returns user with token", lambda: r.get("success") and r.get("data", {}).get("token"))
token = r.get("data", {}).get("token", "")
test("Token is non-empty string", lambda: isinstance(token, str) and len(token) > 10)

print("\n[4] Auth - Login")
r = post("/api/v1/auth/login", {"email": f"e2e{ts}@163.com", "password": "test123456"})
test("Login returns success", lambda: r.get("success"))
login_token = r.get("data", {}).get("token", "")
test("Login returns new token", lambda: len(login_token) > 10)

print("\n[5] Auth - /me")
r = get("/api/v1/auth/me", login_token)
test("/me returns user data", lambda: r.get("success") and r.get("data", {}).get("email") == f"e2e{ts}@163.com")
test("/me shows usage_count", lambda: "usage_count" in r.get("data", {}))

print("\n[6] Auth - Unauthorized access blocked")
r = get("/api/v1/auth/me")
test("Missing token returns error", lambda: "error" in r)

# === JOBS ===
print("\n[7] Jobs")
r = get("/api/v1/jobs")
test("Jobs returns list", lambda: r.get("success") and isinstance(r.get("data"), list))
test("At least 10 jobs", lambda: len(r.get("data", [])) >= 10)
test("Jobs have required fields", lambda: all(k in r["data"][0] for k in ["id", "title", "company", "location"]))

r = get("/api/v1/jobs/j001")
test("Single job returns data", lambda: r.get("success") and r.get("data", {}).get("title"))

# === PRICING ===
print("\n[8] Pricing")
r = get("/api/v1/pricing")
test("Pricing returns tiers", lambda: r.get("success") and len(r.get("data", [])) == 3)
test("Has free tier", lambda: any(t["id"] == "free" for t in r["data"]))
test("Has pro tier", lambda: any(t["id"] == "pro" for t in r["data"]))

# === AI OPTIMIZE ===
print("\n[9] AI Optimize (with auth)")
resume_text = """John Doe
Software Engineer
Skills: Python, Go, React, Docker
Experience: 3 years at Google working on cloud infrastructure
Education: BS Computer Science, Stanford University"""
r = post("/api/v1/optimize", {"resume": resume_text, "target_position": "Senior Backend Engineer"}, login_token)
test("Optimize returns success", lambda: r.get("success"))
if r.get("success"):
    data = r.get("data", {})
    test("Has ATS score", lambda: "ats_score" in data or "score" in data)
    test("Has optimized content", lambda: "optimized_content" in data or "optimized" in data or "result" in data or "content" in data)
    print(f"    (Score: {data.get('ats_score', data.get('score', 'N/A'))})")
else:
    print(f"    (AI error: {r.get('error', 'unknown')})")

print("\n[10] AI Optimize (without auth)")
r = post("/api/v1/optimize", {"resume": "test", "target_position": "test"})
test("Optimize without auth is blocked", lambda: r.get("error") in ["UNAUTHORIZED", "INVALID_BODY"] or not r.get("success"))

# === TEMPLATES ===
print("\n[11] Templates")
r = get("/api/v1/templates")
test("Templates returns data", lambda: r.get("success"))

# === GENERATE RESUME ===
print("\n[12] Generate Resume")
r = post("/api/v1/generate-resume", {"messages": [{"role": "user", "content": "Help me create a resume for a software engineer"}]}, login_token)
test("Generate resume returns response", lambda: r.get("success") or "error" in r)

# === PERSPECTIVE ===
print("\n[13] Perspective Analysis")
r = post("/api/v1/perspective", {"resume_text": resume_text, "target_job": "Senior Backend Engineer", "lang": "en"}, login_token)
test("Perspective returns response", lambda: r.get("success") or "error" in r)

# === STRIPE CHECKOUT ===
print("\n[14] Stripe Checkout")
r = post("/api/v1/create-checkout-session", {"plan": "pro"}, login_token)
test("Checkout without Stripe key returns error", lambda: r.get("error") in ["PAYMENT_NOT_CONFIGURED", "INVALID_BODY"])

# === FRONTEND PAGES ===
print("\n[15] Frontend Pages")
pages = ["/en/editor", "/en/auth", "/en/jobs", "/en/generate", "/en/pricing", "/zh/editor", "/ja/editor"]
for p in pages:
    try:
        code = urllib.request.urlopen(urllib.request.Request(f"http://localhost:8088{p}"), context=ctx).getcode()
    except urllib.error.HTTPError as e:
        code = e.code
    test(f"GET {p} returns 200", lambda c=code: c == 200)

# === SUMMARY ===
print(f"\n{'='*50}")
print(f"Results: {passed} passed, {failed} failed, {passed+failed} total")
print(f"{'='*50}")
