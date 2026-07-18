import pytest
from unittest.mock import AsyncMock, patch

from app.services import risk_engine


@pytest.mark.parametrize("score,expected_level", [
    (0, "low"),
    (39, "low"),
    (40, "medium"),   # exact boundary — must round up to medium, not low
    (74, "medium"),
    (75, "high"),     # exact boundary — must round up to high, not medium
    (100, "high"),
])
def test_level_from_score_thresholds(score, expected_level):
    assert risk_engine._level_from_score(score) == expected_level


@pytest.mark.asyncio
async def test_evaluate_plan_no_changes_returns_zero_risk():
    plan = {"resource_changes": []}
    result = await risk_engine.evaluate_plan(repo_name="test-repo", plan_json=plan)

    assert result.risk_score == 0
    assert result.risk_level == "low"
    assert result.resource_changes == []


@pytest.mark.asyncio
async def test_evaluate_plan_uses_score_derived_level_not_llm_claim():
    """
    If the LLM's stated risk_level disagrees with what its own numeric score
    implies, the deterministic threshold must win. This is the core safety
    guarantee of the risk engine — it should never blindly trust the model's
    self-reported label.
    """
    plan = {
        "resource_changes": [
            {"address": "aws_instance.new", "type": "aws_instance",
             "provider_name": "aws", "change": {"actions": ["create"]}},
        ]
    }

    fake_llm_response = {
        "risk_score": 90,
        "risk_level": "low",  # deliberately wrong/inconsistent with the score
        "reasoning": "test reasoning",
        "flagged_resources": ["aws_instance.new"],
    }

    with patch("app.services.risk_engine.llm_client.score_plan", new=AsyncMock(return_value=fake_llm_response)):
        result = await risk_engine.evaluate_plan(repo_name="test-repo", plan_json=plan)

    assert result.risk_score == 90
    assert result.risk_level == "high"  # derived from score, not the LLM's mislabeled "low"


@pytest.mark.asyncio
async def test_evaluate_plan_passes_through_flagged_resources():
    plan = {
        "resource_changes": [
            {"address": "aws_iam_role.admin", "type": "aws_iam_role",
             "provider_name": "aws", "change": {"actions": ["update"]}},
        ]
    }

    fake_llm_response = {
        "risk_score": 60,
        "risk_level": "medium",
        "reasoning": "IAM role permissions modified",
        "flagged_resources": ["aws_iam_role.admin"],
    }

    with patch("app.services.risk_engine.llm_client.score_plan", new=AsyncMock(return_value=fake_llm_response)):
        result = await risk_engine.evaluate_plan(repo_name="test-repo", plan_json=plan)

    assert result.flagged_resources == ["aws_iam_role.admin"]
    assert result.reasoning == "IAM role permissions modified"