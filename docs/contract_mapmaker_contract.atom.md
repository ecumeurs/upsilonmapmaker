---
id: contract_mapmaker_contract
status: STABLE
type: CONTRACT
dependents:
  - [[rule_mapmaker_board_generation_constraints]]
  - [[rule_mapmaker_seed_determinism]]
layer: BUSINESS
version: 1.0
priority: 1
tags: [governance, contract, mapmaker]
parents:
  - [[shared:contract_upsilon_contract]]
human_name: UpsilonMapMaker Contract
---

# UpsilonMapMaker Contract

## INTENT
Establish the algorithmic constraints and output standards for procedural map generation.

## THE RULE / LOGIC
- **Output Format:** Must produce valid `[[upsilonmapdata]]` structures.
- **Determinism:** Algorithms must be seed-based to allow for match replayability and debugging.
- **Constraints:**
  - Board size: 5-15 tiles per dimension.
  - Minimum area: 50 tiles.
  - Obstacle density: Maximum 10% of total tiles.
- **Pathfinding:** Must verify that at least one path exists between opposing team spawn points.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[contract_mapmaker_contract]]`
- **Related Atoms:** `[[shared:contract_upsilon_contract]]`

## EXPECTATION
