from typing import List, Optional
from pydantic import BaseModel, Field


class ResourceChange(BaseModel):
    address: str
    action: str  # "create" | "update" | "delete" | "replace"
    resource_type: str
    provider: str
    tags: dict = Field(default_factory=dict)


class ScanRequest(BaseModel):
    repo_name: str
    plan_json: dict = Field(..., description="Raw `terraform show -json` plan output")


class ScanResponse(BaseModel):
    repo_name: str
    risk_score: int
    risk_level: str
    reasoning: str
    flagged_resources: List[str]
    resource_changes: List[ResourceChange]
    summary: str
    policy_flags: List[str] = []


class LLMRiskResult(BaseModel):
    risk_score: int
    risk_level: str
    reasoning: str
    flagged_resources: Optional[List[str]] = []