from fastapi import APIRouter, HTTPException

from app.models.schemas import ScanRequest, ScanResponse
from app.services import risk_engine

router = APIRouter(prefix="/scan", tags=["scan"])


@router.post("", response_model=ScanResponse)
async def scan_plan(request: ScanRequest):
    """Accepts a raw `terraform show -json` plan and returns a risk assessment."""
    try:
        result = await risk_engine.evaluate_plan(
            repo_name=request.repo_name,
            plan_json=request.plan_json,
        )
        return result
    except Exception as exc:
        raise HTTPException(status_code=500, detail=f"risk scoring failed: {exc}")