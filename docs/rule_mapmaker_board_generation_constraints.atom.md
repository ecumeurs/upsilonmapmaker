---
id: rule_mapmaker_board_generation_constraints
status: DRAFT
human_name: "MapMaker Generation Constraints"
type: RULE
tags: [mapmaker,generation,constraints]
dependents: []
layer: BUSINESS
version: 1.0
priority: 2
parents:
  - [[contract_mapmaker_contract]]
---

# MapMaker Generation Constraints

## INTENT
Ensure generated maps are tactical and manageable within the engine's performance limits.

## THE RULE / LOGIC
- **Size:** 5 <= dimension <= 15.
- **Area:** Total walkable tiles >= 50.
- **Obstacles:** Density must not exceed 10% of total tile count.
- **Verticality:** Must support at least one ground level.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_mapmaker_board_generation_constraints]]`

## EXPECTATION
Every generated map must fall within the defined size and density ranges.
