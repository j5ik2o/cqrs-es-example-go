#!/usr/bin/env bash
# Pick the commit that PR CI checks: a merge of the current base branch tip
# and the current PR head. Event payloads can carry an old base SHA, and
# refs/pull/N/merge can keep an old base for minutes after the base moves, so
# both are checked against the current PR. When GitHub's merge stays old, this
# builds the merge locally and bundles it for the product jobs.
#
# Run inside a full clone. Inputs (environment): PR_NUMBER, EVENT_HEAD_SHA,
# GH_TOKEN, GITHUB_API_URL, GITHUB_REPOSITORY, GITHUB_OUTPUT, RUNNER_TEMP.
# Outputs: sha, checkout_ref, test_merge ("sha base head"), bundle_name.
set -euo pipefail

attempts=${PR_TEST_MERGE_ATTEMPTS:-6}
sha40='^[0-9a-f]{40}$'
if [[ ! "$PR_NUMBER" =~ ^[1-9][0-9]*$ || ! "$EVENT_HEAD_SHA" =~ $sha40 ]]; then
  echo '::error::Missing or invalid PR number or head SHA in the event'
  exit 1
fi

for (( attempt = 1; attempt <= attempts; attempt++ )); do
  if (( attempt > 1 )); then sleep 5; fi
  # Reading the PR also asks GitHub to recompute refs/pull/N/merge.
  current=$(gh api "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/pulls/$PR_NUMBER" \
    --jq '[.state, .base.ref, .head.sha] | @tsv')
  IFS=$'\t' read -r state base_ref head_sha <<< "$current"
  if [[ "$state" != open || "$head_sha" != "$EVENT_HEAD_SHA" ]]; then
    printf '::error::This event is for head %s, but the PR is %s with head %s\n' \
      "$EVENT_HEAD_SHA" "$state" "$head_sha"
    exit 1
  fi
  # The API's base SHA can lag a push to the base branch; read the tip itself.
  tip=$(git ls-remote --exit-code origin "refs/heads/$base_ref")
  IFS=$'\t' read -r base_sha tip_ref <<< "$tip"
  if [[ ! "$base_sha" =~ $sha40 || "$tip_ref" != "refs/heads/$base_ref" ]]; then
    printf '::error::Cannot read the tip of base branch %s\n' "$base_ref"
    exit 1
  fi
  expected="$base_sha $head_sha"
  parents=
  if git fetch --quiet --no-tags origin "+refs/pull/$PR_NUMBER/merge:refs/pr-ci/candidate"; then
    candidate=$(git rev-parse --verify 'refs/pr-ci/candidate^{commit}')
    parents=$(git show -s --format=%P "$candidate")
  fi
  printf 'Attempt %s: current parents (%s) %s, GitHub merge parents %s\n' \
    "$attempt" "$base_ref" "$expected" "${parents:-none}"
  if [[ "$parents" == "$expected" ]]; then
    printf 'Using GitHub test merge %s\n' "$candidate"
    printf 'sha=%s\ncheckout_ref=%s\ntest_merge=%s %s\n' \
      "$candidate" "$candidate" "$candidate" "$expected" >> "$GITHUB_OUTPUT"
    exit 0
  fi
done

# Build the merge from the parents read in the last attempt, so every product
# job checks the same commit even if a branch moves again.
git fetch --quiet --no-tags origin "$base_sha" "$head_sha"
if ! tree=$(git merge-tree --write-tree "$base_sha" "$head_sha"); then
  printf '::error::Cannot create a conflict-free test merge of %s\n' "$expected"
  exit 1
fi
merge=$(GIT_AUTHOR_NAME='PR CI' GIT_AUTHOR_EMAIL='noreply@github.com' \
  GIT_COMMITTER_NAME='PR CI' GIT_COMMITTER_EMAIL='noreply@github.com' \
  git commit-tree "$tree" -p "$base_sha" -p "$head_sha" -m "Test merge of $head_sha into $base_ref")
git update-ref refs/pr-ci/merge "$merge"
git bundle create --quiet "$RUNNER_TEMP/pr-test-merge.bundle" refs/pr-ci/merge "^$base_sha"
printf 'GitHub merge stayed outdated; built test merge %s\n' "$merge"
printf 'sha=%s\ncheckout_ref=%s\ntest_merge=%s %s\nbundle_name=pr-test-merge-%s\n' \
  "$merge" "$base_sha" "$merge" "$expected" "${GITHUB_RUN_ATTEMPT:-1}" >> "$GITHUB_OUTPUT"
