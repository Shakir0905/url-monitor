# Security Policy

## Supported Versions

This is a personal/portfolio project. Only the latest `main` branch receives security updates.

## Reporting a Vulnerability

If you discover a security vulnerability, please email the maintainer directly rather than opening a public issue:

- ramazanovshakir9@gmail.com
- Telegram: @Shakir_age

Include:
- A description of the vulnerability
- Steps to reproduce
- Potential impact

You can expect a response within a few days.

## Security Best Practices Used

- JWT tokens with HMAC-SHA256, configurable TTL (24h default)
- Bcrypt password hashing
- User-enumeration protection (same error for invalid email and wrong password)
- Required environment variables fail-fast on startup
- Non-root user (UID 1000) in Docker containers
- Secrets via Kubernetes Secrets / Docker env vars, never committed
