---
id: mechanic_mapmaker_seed_determinism
status: DRAFT
human_name: "MapMaker Seed Determinism"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
dependents: []
priority: 2
tags: [mapmaker,determinism,seed]
parents:
  - [[shared:contract_mapmaker_contract]]
---

# New Atom

## INTENT
Enable match replayability and debugging by ensuring map generation is deterministic.

## THE RULE / LOGIC
- **Seed Input:** All random decisions in the generator must be derived from a provided seed.
- **Library Isolation:** Use local random number generators rather than global state to prevent interference from other concurrent processes.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[mapmaker_seed_determinism]]`

## EXPECTATION
Providing the same seed to the generator must produce identical grid results across different execution environments.
