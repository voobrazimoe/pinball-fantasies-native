# Fresh public snapshot procedure

The private development repository stays private. The owner authorized the first
public beta, selected MIT / copyright 2026 voobrazimoe, and approved source-guided
native implementation/audit material. Do not reopen historical manual review gates.
Original commercial payload and personal builds must never be published.

1. Pin the final private-main SHA after license/documentation changes. Review all
   tracked paths and prepare an approval JSON with `commit`, `legal_decisions`
   recording the actual owner decisions, `files` (path to SHA256), and `excluded`
   (path to concrete reason). Exclude private publication manifests/internal notes,
   local logs and personal artifact records. Keep native code, tests, reconstruction
   reports and required build tools. Read and audit every selected blob.
2. Run `python3 tools/export_public_snapshot.py --approval /path/approval.json
   --output /path/new-absent-directory`. The exporter uses only pinned Git blobs,
   checks every path/hash/type and never copies .git, ignored files or symlinks.
3. Run content checks and clean asset-free tests/builds with no originals,
   historical source, developer paths or inherited Go/module caches. Build public
   EXE/AppImage, inspect their content, compare builds with/without local originals
   and generate SHA256SUMS.txt. Personal outputs stay local/private.
4. In the audited export, `git init`, `git branch -M main`, `git add .`, and
   `git commit -m "Initial public release"`. Verify exactly one commit, zero
   parents, and a root tree matching every exported hash/mode. Add only the public
   destination and push only `HEAD:refs/heads/main`.
5. Require green public Linux/Windows CI. Tag the exact verified public release
   commit `v0.1.0-beta` and upload only the public EXE, AppImage and SHA256SUMS.txt
   to a prerelease. Independently verify public history, release and clean clone.
6. Record public identifiers/hashes internally in private development. Never
   merge public Git history back into private history; transfer future content
   fixes without joining the two histories.

Never copy private .git, mirror private history, push --all/--mirror/--tags,
transfer private refs/stashes/bundles or reuse any private commit as a public
parent. Validation clones containing private history must never become public.
