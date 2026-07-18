"""
Thin wrapper around the LLM call used for risk scoring.
Keeping this isolated means swapping providers (Anthropic, OpenAI, local
model) only touches this one file.
"""

import json
import httpx
from app.config import settings

SYSTEM_PROMPT = """You are TerraSentry's infrastructure risk analyst. You review \
Terraform plan changes and score them for operational risk before they are applied.

Score risk based on:
- Destructive actions (delete, replace) on stateful resources (databases, volumes, \
storage buckets) are HIGH risk.
- Security-relevant changes (IAM roles/policies, security groups, network ACLs, \
public access settings) are HIGH risk.
- Changes to production-tagged or prod-named resources raise risk one tier.
- Pure additive changes (create-only, no deletes) to non-critical resources are LOW risk.
- Scaling/config updates to compute resources are MEDIUM risk by default.

Respond ONLY with valid JSON, no markdown fences, no preamble, matching exactly:
{
  "risk_score": <integer 0-100>,
  "risk_level": "<low|medium|high>",
  "reasoning": "<2-3 sentence explanation>",
  "flagged_resources": ["<resource address>", ...]
}"""


async def score_plan(plan_summary: str, repo_name: str) -> dict:
    """Calls the LLM and returns a parsed risk assessment dict."""
    user_prompt = f"Repository: {repo_name}\n\nTerraform plan changes:\n{plan_summary}"

    async with httpx.AsyncClient(timeout=30.0) as client:
        response = await client.post(
            "https://api.anthropic.com/v1/messages",
            headers={
                "x-api-key": settings.LLM_API_KEY,
                "anthropic-version": "2023-06-01",
                "content-type": "application/json",
            },
            json={
                "model": settings.LLM_MODEL,
                "max_tokens": 500,
                "system": SYSTEM_PROMPT,
                "messages": [{"role": "user", "content": user_prompt}],
            },
        )
        response.raise_for_status()
        data = response.json()

    raw_text = "".join(
        block["text"] for block in data.get("content", []) if block.get("type") == "text"
    ).strip()

    # Defensive: strip accidental code fences if the model adds them anyway
    if raw_text.startswith("```"):
        raw_text = raw_text.strip("`")
        raw_text = raw_text.replace("json\n", "", 1).strip()

    try:
        return json.loads(raw_text)
    except json.JSONDecodeError:
        # Fail safe: if the model output isn't parseable, don't silently
        # pass a change through — force human review.
        return {
            "risk_score": 100,
            "risk_level": "high",
            "reasoning": "LLM response could not be parsed; flagged for manual review as a safety default.",
            "flagged_resources": [],
        }