# Migration Verification Plan

## Dual-Run Architecture

The Node.js worker (kealgo) and Python worker (corekg-pipeline) can run simultaneously:

```
Go keparser → Redis Streams → Go keworker (HTTP → Python)  [existing]
Go keparser → NATS JetStream → Node.js kealgo worker       [new]
```

## Verification Steps

1. Deploy kealgo worker in test environment
2. Submit test documents through normal flow
3. Compare chunk outputs:
   - Same document
   - Same split_config
   - Compare: chunk count, chunk text, embedding dimensions
4. Validate ES documents match expected schema

## Switching Traffic

To route tasks to kealgo instead of Python:
1. Ensure NATS bridge is active in keparser
2. Start kealgo worker
3. Monitor task completion rates
4. Gradually increase NATS traffic
5. Decommission Python worker when stable

## Rollback

If issues found:
1. Stop kealgo worker
2. Tasks remain in NATS stream (unack'd)
3. Python worker continues via HTTP polling
4. No data loss
