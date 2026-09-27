# Coordinated catalog additions

The catalog package (and, in api-engine, its SQL repository/migration) is copied
from the sibling first-party repositories for this unreleased coordinated
change. Existing Go module pins remain unchanged until those repositories have
published versions. No third-party dependency source is modified.

`nibble-local-dev/scripts/sync-catalog-vendor.py` reproduces these additions from
the matching sibling checkouts. Standalone CI uses `-mod=vendor`; local Compose
uses sibling Go workspaces. After publishing the model/store modules, bump the
pins and regenerate vendor normally, then remove this temporary bridge.
