package middleware

// Domain failures are mapped in httpx.WriteError. Controllers call that helper
// instead of setting status themselves — the Express-style next(err) analog.
