#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../backend"

# FR-6 acceptance gate: render a certificate through a real Gotenberg, not the
# fake renderer the unit tests use. Kept out of backend.sh so the main suite is
# not slowed by pulling the Chromium image. The test starts its own Gotenberg
# container; set GOTENBERG_URL to point it at an existing instance instead.
go test -tags gotenberg_integration -run TestCertificateRender_RealGotenberg -count=1 -v ./internal/service/
