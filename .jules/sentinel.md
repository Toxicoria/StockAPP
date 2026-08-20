## 2024-08-20 - Missing Authentication on Admin Endpoints
**Vulnerability:** The proxy bypassed authentication checks for all `/api/admin` routes to allow the `front-admin` to log in, but `stock-operations` backend failed to add authentication middleware to any of its other admin routes (`/api/admin/negocios`, etc.), leaving them completely unauthenticated and accessible to anyone.
**Learning:** When whitelisting prefixes at an API gateway/proxy layer, it's easy to accidentally expose internal routes if the downstream service assumes the proxy handled all auth.
**Prevention:** Ensure explicit authentication middleware is applied on every sensitive route at the service level, even if a proxy is present. Don't rely solely on proxy path prefixes for security.
