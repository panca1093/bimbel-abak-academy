# MR review — sha-5ef4997

**Generated:** 2026-07-16 19:47:56 WIB
**Worktree:** /Users/Panca/Documents/MyBook/Project/akademi-bimbel/app
**Base:** main (11e1dab)
**Head:** feat/store-physical-products (5ef4997) — GitHub PR #36
**MR type:** mixed
**Verdict:** **request-changes**

MR-type evidence (`git log main..HEAD --format=%s`): 7× `feat:`, 2× `fix:`, 1× `test:`,
1× `docs:`, 1× `style:`. Feature and fix signals disagree, so per
`diff-heuristics.md` § Scoping rules the MR is classified **mixed** and each file is
judged at the strictest applicable severity. Both `fix:` commits (`f92ece3`, `38f1759`)
repair defects introduced earlier *on this same branch*, not on `main`, and both ship
test files in the same diff — so MR-COV-07 does not fire.

## Summary

| Severity | Count |
|----------|-------|
| Blocker          | 0 |
| Request-changes  | 6 |
| Nit              | 3 |

Changed files: 12 Go files (5 production, 7 test), 16 non-Go (2 SQL migrations, 1 markdown, 13 TypeScript).
Build gate: ✅ passed (`go build ./...` and `go vet ./...` clean)
Targeted tests: ✅ passed — `go test -race -shuffle=on ./internal/{service,handler,worker,repository}/...` and `go test -race -count=1 ./integration/...` all green.

## Blockers

None.

## Request-changes

### R-01 [MR-COV-02] backend/internal/service/store.go:236 (coverage) — `UpdateProductWithCourses` is falsely tested by a shim that encodes the reverted behavior

**Evidence:**

The MR replaces the unconditional field overwrite with a presence-flag branch:

```go
// backend/internal/service/store.go:236 (added)
if !p.WeightGramsSet && p.WeightGrams == 0 {
    p.WeightGrams = existing.WeightGrams
}
```

The only tests naming this method re-declare it on a shim struct rather than calling
the real `*Service`:

```go
// backend/internal/service/store_test.go:1181 (pre-existing, unchanged by this MR)
func (s *shimUpdateProductWithCourses) UpdateProductWithCourses(...) (model.Product, error) {
    ...
    // Preserve non-editable fields from existing record (Bug C fix)
    p.Type = existing.Type
    p.WeightGrams = existing.WeightGrams   // store_test.go:1193 — OLD unconditional overwrite
    p.ImageURL = existing.ImageURL         // store_test.go:1194
```

`store_test.go:1327` calls `svc.UpdateProductWithCourses` where `svc` is
`&shimUpdateProductWithCoursesAtomic{}` (`store_test.go:1325`) — never the real
service. The shim body still contains the exact two lines that commit `f92ece3`
removed from production. The test suite therefore asserts the *old* semantics while
production implements the new ones, and will stay green no matter what the real branch
does.

This is the shim-copy anti-pattern already logged as review debt from PR #25 and
PR #32. This MR did not introduce the shim, but it is the change that makes the shim
actively wrong.

**Suggested fix:**

Add a real-service case in `backend/integration/update_preserve_test.go` calling
`env.svc.UpdateProductWithCourses` for both the persist-new and preserve-on-omit
paths. Separately, realign or delete the stale shim at `store_test.go:1181-1194` so it
no longer encodes reverted behavior.

---

### R-02 [MR-COV-02] backend/internal/service/store.go:181 (coverage) — `UpdateProductWithExams` got the new preserve branch with zero test references

**Evidence:**

```go
// backend/internal/service/store.go:181 (added)
if !p.WeightGramsSet && p.WeightGrams == 0 {
    p.WeightGrams = existing.WeightGrams
}
// store.go:184
if !p.ImageURLSet && p.ImageURL == "" {
    p.ImageURL = existing.ImageURL
}
```

`grep -rn 'UpdateProductWithExams' --include='*_test.go'` returns **zero hits**
module-wide (verified). The path is live in production: `AdminUpdateProduct` routes to
it whenever `exam_ids` is non-nil (`backend/internal/handler/product.go:196`). The
sibling copy in `UpdateProduct` is covered by
`backend/integration/update_preserve_test.go:87`; this copy is protected by nothing.

**Suggested fix:**

Add a case alongside `TestUpdateProduct_PersistsNewWeightImage_ThenPreservesOnOmit_RealService`
in `backend/integration/update_preserve_test.go` that calls
`env.svc.UpdateProductWithExams` with (a) a new `WeightGrams`+`ImageURL`, asserting both
persist, and (b) a follow-up call omitting both, asserting the persisted values survive.

---

### R-03 [MR-QUAL-03] backend/internal/service/store_merch_test.go:86 (coverage) — ship-before-complete gate is tested in one direction only

**Evidence:**

```go
// backend/internal/service/store_merch_test.go:86
func TestAdminCompleteOrder_ProcessingMerchandise_RejectedUntilShipped(t *testing.T) {
    ...
    require.NoError(t, repo.SetOrderStatus(ctx, tx, order.ID, "processing", ""))
    ...
    err = svc.AdminCompleteOrder(ctx, order.ID.String())
    if err == nil || !strings.Contains(err.Error(), "must be shipped before completing") {
        t.Errorf(...)
    }
}   // store_merch_test.go:107 — test ends here; "until shipped" is never exercised
```

The name promises "RejectedUntilShipped" but the test never drives the order to
`shipped` and asserts that completion then succeeds. A SUT that rejects
`AdminCompleteOrder` for *every* order — physical or digital — passes this test. The
gate at `store.go:989` is the core of the feature, and the new `isPhysicalType()`
widening feeding it could be inverted to `!isPhysicalType` with this test still green.

**Suggested fix:**

Add a second subtest: after `SetOrderStatus(order, "shipped")`, assert
`require.NoError(t, svc.AdminCompleteOrder(ctx, order.ID.String()))` and that the status
becomes `completed`. Optionally add a digital-only (course/exam) order asserting it
completes from `processing` without shipping, pinning that the gate is type-scoped
rather than blanket.

---

### R-04 [MR-COV-06] backend/internal/service/store.go:294 (coverage) — the `ImageURLSet` explicit-clear branch is unreached

**Evidence:**

```go
// backend/internal/service/store.go:294 (added)
if !p.ImageURLSet && p.ImageURL == "" {
    p.ImageURL = existing.ImageURL
}
```

The flag introduces a branch the old unconditional `p.ImageURL = existing.ImageURL`
could not express: an explicit `image_url: ""` (flag true, value empty) skips the
preserve and **clears the stored image**. `grep -rn 'ImageURLSet\|WeightGramsSet'
--include='*_test.go'` returns zero hits module-wide, and no test sends an empty
`image_url`. The symmetric weight case *is* covered — `product_merch_handler_test.go:172`
asserts explicit `weight_grams: 0` persists — which makes the image omission
conspicuous, since both flags exist for the same reason.

**Suggested fix:**

Add a handler case mirroring `TestAdminUpdateProduct_WeightGramsZero_PersistsAndPreservesOmittedImage`:
PATCH `"image_url": ""` on a product with a stored image, assert response and persisted
row both show an empty `image_url`, and assert `weight_grams` is preserved. If clearing
is *not* intended, the branch should reject empty instead — either way a test must pin
the decision.

---

### R-05 [RUBRIC] backend/internal/service/store_merch_test.go:103 (assertions) — sentinel-string error assertion

**Evidence:**

```go
// backend/internal/service/store_merch_test.go:103
if err == nil || !strings.Contains(err.Error(), "must be shipped before completing") {
```

Per `anti-patterns.md` (golang/08 §10) errors must be matched with `errors.Is`/`As` on a
typed sentinel. The SUT builds this error inline via `errors.New` at `store.go:989` with
no exported sentinel, so the test is coupled to prose that any reword — or an i18n pass
— silently breaks. The sibling tests in this same file correctly use `errors.Is`/`ErrorIs`
(lines 24, 32, 68, 81) against `ErrForbidden`/`ErrOutOfStock`/`ErrNotFound`, so this is an
inconsistency within the new code, not a repo-wide constraint.

**Suggested fix:**

Add `ErrMustShipBeforeComplete = errors.New("order has physical items — must be shipped before completing")`
to the sentinel block in `store.go` (already touched by this MR, lines 14-26), return it
at `store.go:989`, and assert `require.ErrorIs(t, err, ErrMustShipBeforeComplete)`.

---

### R-06 [RUBRIC] backend/internal/handler/product_merch_handler_test.go:159 (isolation) — swallowed fixture errors produce a misleading diagnostic

**Evidence:**

```go
// backend/internal/handler/product_merch_handler_test.go:159
tokenString, jti, _ := env.signer.SignAccess("admin_exam_user", service.RoleAdminExam, nil, []string{})
rdb.Set(context.Background(), "session:access:"+jti, "admin_exam_user", 15*time.Minute)   // :160, .Err() unchecked
```

Line 159 discards the `SignAccess` error; line 160 ignores `.Err()`. If either fails the
token is empty/unregistered and the request 401s — the test then fails with
"want 403, got 401", pointing the reader at RBAC when the real fault is fixture setup.
The `mintProductToken` helper at `product_merch_handler_test.go:113` in this same file
checks both; this test bypasses it for no stated reason.

**Severity deviation — flagged for your override.** `review-rubric.md` buckets swallowed
errors as **Blocker**. Both this review and the test-quality agent rated it
request-changes instead, because the governing meta-heuristic ("would a broken
implementation slip past this test?") answers *no* — these are fixture calls, not SUT
calls, so a failure still reddens the test, merely with a misleading message. If your
team reads that rule strictly, promote this to Blocker.

**Suggested fix:**

Reuse the existing helper shape: `tokenString, jti, err := env.signer.SignAccess(...)`
with `if err != nil { t.Fatalf("SignAccess: %v", err) }`, and
`if err := rdb.Set(...).Err(); err != nil { t.Fatalf("redis set session: %v", err) }`.

---

## Nits

- [N-01] backend/internal/handler/product_merch_handler_test.go:42 (maintainability) — `t.Fatalf` inside `adminProductDBOnce.Do` (lines 53, 57, 60, 64, 70) triggers `runtime.Goexit` on the first test's goroutine, marking the `Once` done while leaving `adminProductDBEnv` nil; every later test then reports the generic "admin product test env failed to initialize" (line 108) instead of the real container error, and under `-shuffle=on` which test carries the true diagnostic varies per run. Capture the error in a package-level var and surface it after the `Do`. (Mirrors the established `realdb_test.go:30` convention, and the container *is* properly terminated in `main_test.go:86` — hence nit.)
- [N-02] backend/internal/repository/order_checkout_merch_test.go:15 (intent) — `TestCheckoutOrder_MerchandiseStockEnforcedAndDecremented` drives two behaviors (sufficient-stock decrement, lines 21-27; insufficient-stock rejection with stock unchanged, lines 30-37) in one body with no subtests, so a failure in the first half aborts before the second runs. Split into two `t.Run` subtests. (Not true assertion roulette — every `require` carries an explanatory message.)
- [N-03] backend/internal/worker/outbox_merch_test.go:39 (isolation, MR-QUAL-04) — `time.Now().String()` on an added line for the `OutboxEvent.CreatedAt` fixture. Not a flake source (`CreatedAt` is never asserted on and `pollOutbox` branches nothing on it), but a frozen literal would read as deliberately fixed rather than incidentally time-dependent.

## What this MR got right

Recorded because it is load-bearing for the verdict — these are the checks that would
normally produce Blockers and did not:

- **MR-COV-05 (SQL/migration coverage) is satisfied, and is the strongest part of the MR.** `backend/internal/repository/migration_0027_test.go` runs the real migration set against real Postgres (testcontainers `postgres:16-alpine` via `newMigration0025Pool`): pre-migration rejection, post-up acceptance for **both** `merchandise` and `medal`, a `'bogus'` negative control proving the widened CHECK still constrains (this is what stops the migration being written as a CHECK *drop*), fail-safe down conversion to `book`, and post-down re-rejection. No `sqlmock`, no SQLite anywhere in the module.
- **No shim-copy tests among the new files.** Every new test exercises real production code — real echo handlers + real service + real Postgres at the handler layer, real `w.pollOutbox` at the worker layer. (The stale shim in R-01 is pre-existing.)
- **No tautologies.** Every expectation is a hand-written literal; `outbox_merch_test.go:19` iterates a literal `[]string{"merchandise", "medal"}` declared in the test, not read from `isPhysicalType`.
- **The medal tests genuinely discriminate.** `repository/order.go:404` spells the guard as three separate literals and `worker/outbox.go:227` as `case "book", "merchandise", "medal"` — deleting the medal literal from either is caught only by the medal test. They are not copy-paste duplicates.
- **The explicit-zero vs omitted semantics are correctly pinned by a non-contradictory pair:** `product_merch_handler_test.go:172` sends `weight_grams: 0` as an explicit JSON value (`*int` non-nil → `WeightGramsSet=true`, 0 persists), while `integration/update_preserve_test.go` omits the field at the service layer (`Set=false`, zero value → 350 preserved).

## What was not reviewed

- **The 13 `web/` TypeScript files** (`ProductModal.tsx`, `ProductCard.tsx`, catalog and admin pages, `types.ts`, `i18n.ts`) and their Vitest tests. This skill is Go-scoped; the frontend is covered by the separate line-by-line review accompanying this run.
- `docs/superpowers/specs/2026-07-16-medal-product-type-design.md` — not Go.
- MR-COV-01, MR-COV-03, MR-COV-04 — no new non-test `.go` files, no new exported funcs/types/vars/consts, and no new packages in the diff. The new exported *fields* `model.Product.WeightGramsSet`/`ImageURLSet` (`backend/internal/model/product.go:14-15`) fall outside MR-COV-03's detection regex; their gap is reported under R-04 rather than double-counted.
- MR-QUAL-05 — no test-only package changes; every touched test package also has production changes.

**Convention deviation (informational, not a finding):** this repo has **zero**
`//go:build integration` tags module-wide — testcontainers tests run inside the default
`go test` suite by design. MR-COV-05's literal build-tag requirement is met in substance
(real Postgres) but not in form. Pre-existing repo-wide convention, not introduced here.

**Structural risk (informational, raised by both agents, outside test scope):** the
physical-type list is now duplicated across three packages — `service.isPhysicalType`
(`store.go:1199`), `repository/order.go:404`, and `worker/outbox.go:227` — each with its
own inline literals; `order.go:399` documents this as deliberate to avoid a repo→service
import. No test pins the three lists in agreement, so a future fourth physical type added
to one but not the others would ship silently. Not a finding against the tests in this
diff; it is the structural risk the merch/medal tests are compensating for.

## Next steps

- Verdict is `request-changes`: address the six Request-changes findings; Nits at author's discretion.
- R-01 and R-03 are the two worth doing before merge: R-01 because a green test currently asserts reverted behavior, R-03 because the feature's core gate is half-tested.
- When addressing R-01/R-02, the followup is `/ralp-test-plan` on `internal/service` — this skill does not write tests.
- This review has not been posted anywhere. Publishing to a PR is a separate explicit step; this skill never approves.
