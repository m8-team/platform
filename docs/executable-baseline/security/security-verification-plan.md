---
title: "Security Verification Plan"
---

# Security Verification Plan

[Executable Baseline](../index.md) | [Security Baseline](index.md)

Gate включает: SAST, dependency and container scan, IaC scan, secret scan,
authentication abuse tests, authorization matrix, replay tests, SSRF/egress policy,
audit minimization, credential rotation и disaster recovery of trust stores.

Privileged changes требуют four-eyes approval. Production access short-lived и audited.
