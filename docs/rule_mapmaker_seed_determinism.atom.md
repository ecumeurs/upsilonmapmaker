---
id: rule_mapmaker_seed_determinism
status: STABLE
priority: 2
version: 1.0
parents:
  - [[contract_mapmaker_contract]]
human_name: "Procedural Seed Determinism"
type: RULE
dependents:
  - [[mechanic_mapmaker_seed_determinism]]
layer: BUSINESS
---

# Procedural Seed Determinism

## INTENT
Ensure that procedural map generation is perfectly deterministic given a specific seed.

## THE RULE / LOGIC
- Every random choice in the generation algorithm must be derived from the provided seed.
- Sequential calls with the same seed must produce bit-identical grid structures.
- Use the shared `upsilontools/tools` package for all randomization.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_mapmaker_seed_determinism]]`
- **Test Names:** `TestSeedDeterminism`

## EXPECTATION
- Generating a board twice with the same seed yields bit-identical grid structures (tiles, obstacles, spawns).
- Two different seeds produce different boards with overwhelming probability.
- `TestSeedDeterminism` passes; no reliance on global/shared RNG state across concurrent generations.
