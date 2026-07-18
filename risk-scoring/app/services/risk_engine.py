"""
Orchestrates parsing + LLM scoring, and applies deterministic threshold
overrides on top of the LLM's judgment so the system doesn't rely on the
model alone for the low/medium/high cutoffs.
"""

from app.config import settings
from app.models.schemas import LLMRiskResult, ScanResponse
from app.services import llm_client, terraform_parser


def _level_from_score(score: int) -> str:
    if score >= settings.RISK_SCORE_HIGH_THRESHOLD:
        return "high"
    if score >= settings.RISK_SCORE_MEDIUM_THRESHOLD:
        return "medium"
    return "low"


async def evaluate_plan(repo_name: str, plan_json: dict) -> ScanResponse:
    changes = terraform_parser.parse_plan(plan_json)
    summary = terraform_parser.build_plan_summary(changes)

    if not changes:
        return ScanResponse(
            repo_name=repo_name,
            risk_score=0,
            risk_level="low",
            reasoning="No effective resource changes in this plan.",
            flagged_resources=[],
            resource_changes=[],
            summary=summary,
        )

    llm_result_raw = await llm_client.score_plan(plan_summary=summary, repo_name=repo_name)
    llm_result = LLMRiskResult(**llm_result_raw)

    # Deterministic threshold re-check: the LLM's stated level must agree with
    # the score-derived tier. If they disagree, trust the score thresholds —
    # keeps behavior predictable and testable rather than fully model-driven.
    resolved_level = _level_from_score(llm_result.risk_score)

    return ScanResponse(
        repo_name=repo_name,
        risk_score=llm_result.risk_score,
        risk_level=resolved_level,
        reasoning=llm_result.reasoning,
        flagged_resources=llm_result.flagged_resources or [],
        resource_changes=changes,
        summary=summary,
    )