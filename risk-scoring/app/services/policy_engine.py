"""
Declarative policy layer that supplements — never replaces — the LLM's risk
judgment. Rules are loaded once from policy.yaml at startup and act as a
floor: each matching rule can only raise the final risk level, never lower
it below what the LLM/threshold logic already decided.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import yaml

from app.models.schemas import ResourceChange

# Ordering used to compare risk levels numerically so "does this rule raise
# the level" is a simple integer comparison instead of string logic.
_LEVEL_RANK = {"low": 0, "medium": 1, "high": 2}
_RANK_TO_LEVEL = {v: k for k, v in _LEVEL_RANK.items()}


@dataclass
class PolicyRule:
    name: str
    min_level: str
    reason: str = ""
    resource_type_prefix: Optional[str] = None
    action: Optional[str] = None
    tag_equals: Dict[str, str] = field(default_factory=dict)

    def matches(self, change: ResourceChange) -> bool:
        if self.resource_type_prefix and not change.resource_type.startswith(self.resource_type_prefix):
            return False
        if self.action and change.action != self.action:
            return False
        for key, expected in self.tag_equals.items():
            if change.tags.get(key) != expected:
                return False
        return True


class PolicyEngine:
    """Holds a set of loaded PolicyRules and applies them as a floor on top
    of an already-computed risk level."""

    def __init__(self, rules: List[PolicyRule]):
        self.rules = rules

    @classmethod
    def load(cls, path: str) -> "PolicyEngine":
        """Load rules from a YAML file. A missing file yields an empty
        (no-op) engine rather than raising, since the policy layer is
        optional — the LLM/threshold scoring still works without it."""
        policy_path = Path(path)
        if not policy_path.exists():
            return cls(rules=[])

        with policy_path.open("r", encoding="utf-8") as f:
            raw = yaml.safe_load(f) or {}

        rules = [
            PolicyRule(
                name=entry.get("name", "unnamed-rule"),
                min_level=entry.get("min_level", "medium"),
                reason=entry.get("reason", ""),
                resource_type_prefix=entry.get("resource_type_prefix"),
                action=entry.get("action"),
                tag_equals=entry.get("tag_equals") or {},
            )
            for entry in raw.get("rules", [])
        ]
        return cls(rules=rules)

    def apply(self, changes: List[ResourceChange], current_level: str) -> Tuple[str, List[str]]:
        """Evaluate every rule against every changed resource. Returns the
        (possibly raised) final level plus a list of human-readable strings
        describing which rules fired and on which resource, for surfacing
        in the scan's reasoning/flags."""
        current_rank = _LEVEL_RANK.get(current_level, 0)
        final_rank = current_rank
        triggered: List[str] = []

        for change in changes:
            for rule in self.rules:
                if not rule.matches(change):
                    continue
                rule_rank = _LEVEL_RANK.get(rule.min_level, 0)
                triggered.append(f"{rule.name}: {change.address} ({rule.reason})")
                if rule_rank > final_rank:
                    final_rank = rule_rank

        final_level = _RANK_TO_LEVEL.get(final_rank, current_level)
        return final_level, triggered