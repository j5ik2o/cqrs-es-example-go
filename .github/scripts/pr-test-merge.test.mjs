// Tests for the PR test merge that every product job in ci.yml checks.
// Run: node --test .github/scripts/pr-test-merge.test.mjs
import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { chmodSync, existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";

const root = new URL("../..", import.meta.url).pathname;
const script = join(root, ".github/scripts/pr-test-merge.sh");
const workflow = readFileSync(join(root, ".github/workflows/ci.yml"), "utf8");

function git(cwd, ...args) {
  return execFileSync("git", ["-C", cwd, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.com", ...args], {
    encoding: "utf8",
  }).trim();
}

// origin holds main, the PR head and GitHub's merge ref; clone is the runner's checkout.
function fixture(t) {
  const dir = mkdtempSync(join(tmpdir(), "pr-test-merge-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  const origin = join(dir, "origin.git");
  const work = join(dir, "work");
  git(dir, "init", "--quiet", "--bare", origin);
  git(dir, "init", "--quiet", work);
  const commit = (message, file, content, ...parents) => {
    git(work, "checkout", "--quiet", "--detach", parents[0] ?? "HEAD");
    writeFileSync(join(work, file), content);
    git(work, "add", file);
    git(work, "commit", "--quiet", "-m", message);
    return git(work, "rev-parse", "HEAD");
  };
  git(work, "commit", "--quiet", "--allow-empty", "-m", "common");
  const common = git(work, "rev-parse", "HEAD");
  const oldBase = commit("old base", "base.txt", "old\n", common);
  const base = commit("base", "base.txt", "new\n", oldBase);
  const head = commit("head", "head.txt", "head\n", common);
  const oldHead = commit("old head", "head.txt", "old head\n", common);
  const conflict = commit("conflict", "base.txt", "conflict\n", common);
  const merge = (b, h) => git(work, "commit-tree", git(work, "merge-tree", "--write-tree", b, h), "-p", b, "-p", h, "-m", "GitHub merge");
  const push = (refs) => {
    for (const [ref, sha] of Object.entries(refs)) git(work, "push", "--quiet", "--force", origin, `${sha}:${ref}`);
  };
  push({ "refs/heads/main": base, "refs/heads/release": oldBase, "refs/pull/1/head": head });
  const runner = join(dir, "runner");
  git(dir, "clone", "--quiet", origin, runner);
  git(runner, "checkout", "--quiet", "--detach", head);

  const bin = join(dir, "bin");
  execFileSync("mkdir", [bin]);
  writeFileSync(join(bin, "gh"), '#!/bin/sh\nprintf \'%s\\n\' "$FIXTURE_PR"\n');
  writeFileSync(join(bin, "sleep"), "#!/bin/sh\n");
  chmodSync(join(bin, "gh"), 0o755);
  chmodSync(join(bin, "sleep"), 0o755);

  const output = join(dir, "output");
  const resolve = ({ pr = `open\tmain\t${head}`, eventHead = head } = {}) => {
    rmSync(output, { force: true });
    const result = spawnSync("bash", [script], {
      cwd: runner,
      encoding: "utf8",
      env: {
        ...process.env,
        PATH: `${bin}:${process.env.PATH}`,
        PR_NUMBER: "1",
        EVENT_HEAD_SHA: eventHead,
        FIXTURE_PR: pr,
        GH_TOKEN: "fixture",
        GITHUB_API_URL: "https://api.invalid",
        GITHUB_REPOSITORY: "fixture",
        GITHUB_OUTPUT: output,
        RUNNER_TEMP: dir,
        GITHUB_RUN_ATTEMPT: "1",
      },
    });
    const outputs = existsSync(output)
      ? Object.fromEntries(readFileSync(output, "utf8").trim().split("\n").map((line) => line.split(/=(.*)/s).slice(0, 2)))
      : null;
    return { ...result, log: result.stdout + result.stderr, outputs };
  };

  // Runs a product job's checkout and verify steps from ci.yml in a fresh clone.
  const productJob = (outputs) => {
    const job = join(dir, `job-${Math.random().toString(36).slice(2)}`);
    git(dir, "clone", "--quiet", "--no-checkout", origin, job);
    // Like actions/checkout, fetch only the tip unless a bundle needs the base history.
    git(job, "fetch", "--quiet", ...(outputs.bundle_name ? [] : ["--depth=1"]), "origin", outputs.checkout_ref);
    git(job, "checkout", "--quiet", "--detach", outputs.checkout_ref);
    const download = join(dir, "pr-test-merge");
    execFileSync("mkdir", ["-p", download]);
    if (outputs.bundle_name) execFileSync("cp", [join(dir, "pr-test-merge.bundle"), download]);
    return spawnSync("bash", ["-e", "-o", "pipefail", "-c", verifyStep()], {
      cwd: job,
      encoding: "utf8",
      env: { ...process.env, TEST_MERGE: outputs.test_merge, BUNDLE: outputs.bundle_name ?? "", RUNNER_TEMP: dir },
    });
  };

  return { origin, base, oldBase, head, oldHead, conflict, merge, push, resolve, productJob };
}

function verifyStep() {
  const step = workflow.split("      - &verify-test-merge\n")[1];
  const lines = step.split("        run: |\n")[1].split("\n");
  const body = lines.slice(0, lines.findIndex((line) => line && !line.startsWith("          ")));
  return body.map((line) => line.slice(10)).join("\n");
}

function parentsOf(outputs) {
  const [, base, head] = outputs.test_merge.split(" ");
  return { base, head };
}

test("uses GitHub's test merge when its parents are the current base tip and head", (t) => {
  const f = fixture(t);
  const candidate = f.merge(f.base, f.head);
  f.push({ "refs/pull/1/merge": candidate });
  const r = f.resolve();
  assert.equal(r.status, 0, r.log);
  assert.deepEqual(r.outputs, { sha: candidate, checkout_ref: candidate, test_merge: `${candidate} ${f.base} ${f.head}` });
  const job = f.productJob(r.outputs);
  assert.equal(job.status, 0, job.stdout + job.stderr);
});

test("ignores the event's base SHA and checks against the current base tip", () => {
  const resolveStep = workflow.split("        id: resolve\n")[1].split("\n      - ")[0];
  assert.doesNotMatch(resolveStep, /base\.sha/);
  assert.doesNotMatch(readFileSync(script, "utf8"), /BASE_SHA/);
});

test("builds and bundles the test merge when GitHub's merge keeps an old base", (t) => {
  const f = fixture(t);
  f.push({ "refs/pull/1/merge": f.merge(f.oldBase, f.head) });
  const r = f.resolve();
  assert.equal(r.status, 0, r.log);
  assert.match(r.log, /Attempt 6: /);
  assert.deepEqual(parentsOf(r.outputs), { base: f.base, head: f.head });
  assert.equal(r.outputs.checkout_ref, f.base);
  assert.equal(r.outputs.bundle_name, "pr-test-merge-1");
  const job = f.productJob(r.outputs);
  assert.equal(job.status, 0, job.stdout + job.stderr);
  assert.match(job.stdout, new RegExp(`${r.outputs.sha} ${f.base} ${f.head}`));
});

test("checks the tip of the PR's current base branch after a retarget", (t) => {
  const f = fixture(t);
  f.push({ "refs/pull/1/merge": f.merge(f.base, f.head) });
  const r = f.resolve({ pr: `open\trelease\t${f.head}` });
  assert.equal(r.status, 0, r.log);
  assert.deepEqual(parentsOf(r.outputs), { base: f.oldBase, head: f.head });
});

test("fails an event whose head is no longer the PR head", (t) => {
  const f = fixture(t);
  f.push({ "refs/pull/1/merge": f.merge(f.base, f.oldHead) });
  const r = f.resolve({ eventHead: f.oldHead });
  assert.notEqual(r.status, 0);
  assert.match(r.log, /This event is for head/);
  assert.equal(r.outputs, null);
});

test("fails when the PR is no longer open", (t) => {
  const f = fixture(t);
  const r = f.resolve({ pr: `closed\tmain\t${f.head}` });
  assert.notEqual(r.status, 0);
  assert.equal(r.outputs, null);
});

test("fails when the current base and head conflict", (t) => {
  // GitHub keeps no merge ref for a conflicting PR; both sides add base.txt.
  const f = fixture(t);
  f.push({ "refs/pull/1/head": f.conflict });
  const r = f.resolve({ pr: `open\tmain\t${f.conflict}`, eventHead: f.conflict });
  assert.notEqual(r.status, 0);
  assert.match(r.log, /Cannot create a conflict-free test merge/);
  assert.equal(r.outputs, null);
});

test("product jobs reject a checkout that is not the resolved test merge", (t) => {
  const f = fixture(t);
  const stale = f.merge(f.oldBase, f.head);
  f.push({ "refs/pull/1/merge": stale });
  const job = f.productJob({ checkout_ref: stale, test_merge: `${stale} ${f.base} ${f.head}` });
  assert.notEqual(job.status, 0);
  assert.match(job.stdout, /::error::Expected test merge/);
});

test("every pull_request event resolves and verifies the test merge, and Required CI always runs", () => {
  assert.match(workflow, /types: \[opened, synchronize, reopened, edited\]/);
  for (const job of ["lint", "test", "e2e", "aislop"]) {
    const body = workflow.split(`\n  ${job}:\n`)[1].split(/\n  [a-z-]+:\n/)[0];
    assert.match(body, /needs: .*checkout-revision/, job);
    assert.match(body, /- (&checkout|\*checkout)\n/, job);
    assert.match(body, /- (&verify-test-merge|\*verify-test-merge)\n/, job);
    assert.doesNotMatch(body, /github\.event\.action/, job);
  }
  const required = workflow.split("\n  required-ci:\n")[1];
  assert.match(required, /name: Required CI/);
  assert.match(required, /needs: \[checkout-revision, lint, test, e2e, aislop\]/);
  assert.match(required, /if: \$\{\{ always\(\) \}\}/);
});
