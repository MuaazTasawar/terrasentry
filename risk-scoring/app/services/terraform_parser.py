"""
Parses `terraform show -json` plan output into a compact, structured
summary the LLM can reason over without needing the full raw plan
(which can be huge and mostly noise).
"""

from app.models.schemas import ResourceChange


def parse_plan(plan_json: dict) -> list[ResourceChange]:
    """Extract resource changes from a Terraform JSON plan."""
    changes = []
    resource_changes = plan_json.get("resource_changes", [])

    for rc in resource_changes:
        actions = rc.get("change", {}).get("actions", [])
        if actions == ["no-op"] or not actions:
            continue

        action = _normalize_actions(actions)

        changes.append(
            ResourceChange(
                address=rc.get("address", "unknown"),
                action=action,
                resource_type=rc.get("type", "unknown"),
                provider=rc.get("provider_name", "unknown"),
                tags=_extract_tags(rc, action),
            )
        )

    return changes


def _extract_tags(rc: dict, action: str) -> dict:
    """Pull the `tags` attribute off a resource change. Deletes have a null
    `after` block, so we read tags from `before` in that case; every other
    action reads from `after` (the state the resource is moving to)."""
    change = rc.get("change", {})
    side = change.get("before") if action == "delete" else change.get("after")
    if not isinstance(side, dict):
        return {}
    tags = side.get("tags")
    return tags if isinstance(tags, dict) else {}


def _normalize_actions(actions: list[str]) -> str:
    if "delete" in actions and "create" in actions:
        return "replace"
    if "delete" in actions:
        return "delete"
    if "create" in actions:
        return "create"
    if "update" in actions:
        return "update"
    return actions[0] if actions else "unknown"


def build_plan_summary(changes: list[ResourceChange]) -> str:
    """Human-readable summary used both for LLM context and DB storage."""
    if not changes:
        return "No resource changes detected in this plan."

    lines = [f"{len(changes)} resource change(s) detected:"]
    for c in changes:
        lines.append(f"  - [{c.action.upper()}] {c.address} ({c.resource_type})")

    return "\n".join(lines)