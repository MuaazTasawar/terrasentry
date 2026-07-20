import textwrap

import pytest

from app.models.schemas import ResourceChange
from app.services.policy_engine import PolicyEngine, PolicyRule


def _change(resource_type="aws_instance", action="create", tags=None, address="test.resource"):
    return ResourceChange(
        address=address,
        action=action,
        resource_type=resource_type,
        provider="aws",
        tags=tags or {},
    )


# --- PolicyRule.matches ---

def test_rule_matches_on_resource_type_prefix():
    rule = PolicyRule(name="iam", min_level="medium", resource_type_prefix="aws_iam_")
    assert rule.matches(_change(resource_type="aws_iam_role")) is True
    assert rule.matches(_change(resource_type="aws_s3_bucket")) is False


def test_rule_matches_on_action():
    rule = PolicyRule(name="deletes", min_level="high", action="delete")
    assert rule.matches(_change(action="delete")) is True
    assert rule.matches(_change(action="create")) is False


def test_rule_matches_on_tag_equals():
    rule = PolicyRule(name="prod", min_level="high", tag_equals={"env": "prod"})
    assert rule.matches(_change(tags={"env": "prod"})) is True
    assert rule.matches(_change(tags={"env": "staging"})) is False
    assert rule.matches(_change(tags={})) is False


def test_rule_matches_requires_all_conditions_combined():
    rule = PolicyRule(name="prod-deletes", min_level="high", action="delete", tag_equals={"env": "prod"})
    assert rule.matches(_change(action="delete", tags={"env": "prod"})) is True
    assert rule.matches(_change(action="create", tags={"env": "prod"})) is False
    assert rule.matches(_change(action="delete", tags={"env": "staging"})) is False


# --- PolicyEngine.apply ---

def test_apply_raises_level_when_rule_matches():
    engine = PolicyEngine(rules=[PolicyRule(name="iam", min_level="medium", resource_type_prefix="aws_iam_")])
    changes = [_change(resource_type="aws_iam_role", address="aws_iam_role.admin")]

    level, flags = engine.apply(changes, current_level="low")

    assert level == "medium"
    assert len(flags) == 1
    assert "iam" in flags[0]
    assert "aws_iam_role.admin" in flags[0]


def test_apply_never_lowers_level_below_current():
    # Regression guard: policy is a floor, not a ceiling.
    engine = PolicyEngine(rules=[PolicyRule(name="iam", min_level="medium", resource_type_prefix="aws_iam_")])
    changes = [_change(resource_type="aws_iam_role")]

    level, _ = engine.apply(changes, current_level="high")

    assert level == "high"


def test_apply_no_matching_rules_leaves_level_unchanged():
    engine = PolicyEngine(rules=[PolicyRule(name="iam", min_level="medium", resource_type_prefix="aws_iam_")])
    changes = [_change(resource_type="aws_s3_bucket")]

    level, flags = engine.apply(changes, current_level="low")

    assert level == "low"
    assert flags == []


def test_apply_multiple_rules_takes_highest():
    engine = PolicyEngine(rules=[
        PolicyRule(name="iam", min_level="medium", resource_type_prefix="aws_iam_"),
        PolicyRule(name="prod-deletes", min_level="high", action="delete", tag_equals={"env": "prod"}),
    ])
    changes = [_change(resource_type="aws_iam_role", action="delete", tags={"env": "prod"})]

    level, flags = engine.apply(changes, current_level="low")

    assert level == "high"
    assert len(flags) == 2


def test_apply_empty_ruleset_is_a_noop():
    engine = PolicyEngine(rules=[])
    level, flags = engine.apply([_change()], current_level="medium")
    assert level == "medium"
    assert flags == []


# --- PolicyEngine.load ---

def test_load_missing_file_returns_empty_engine(tmp_path):
    engine = PolicyEngine.load(str(tmp_path / "does-not-exist.yaml"))
    assert engine.rules == []


def test_load_parses_rules_from_yaml(tmp_path):
    policy_file = tmp_path / "policy.yaml"
    policy_file.write_text(textwrap.dedent("""
        rules:
          - name: iam-changes-at-least-medium
            resource_type_prefix: aws_iam_
            min_level: medium
            reason: elevated blast radius
          - name: prod-deletes-are-high
            action: delete
            tag_equals:
              env: prod
            min_level: high
    """))

    engine = PolicyEngine.load(str(policy_file))

    assert len(engine.rules) == 2
    assert engine.rules[0].name == "iam-changes-at-least-medium"
    assert engine.rules[0].resource_type_prefix == "aws_iam_"
    assert engine.rules[0].reason == "elevated blast radius"
    assert engine.rules[1].tag_equals == {"env": "prod"}


def test_load_empty_file_returns_empty_ruleset(tmp_path):
    policy_file = tmp_path / "empty.yaml"
    policy_file.write_text("")

    engine = PolicyEngine.load(str(policy_file))
    assert engine.rules == []