from fastapi import FastAPI
from app.config import settings
from app.routers import scan

app = FastAPI(
    title="TerraSentry Risk Scoring Service",
    description="Scores Terraform plans for cost, security, and drift risk using an LLM.",
    version="0.1.0",
)

app.include_router(scan.router)

@app.get("/health")
def health_check():
    return {"status": "ok", "env": settings.ENV}


# The /scan router (app/routers/scan.py) is wired in Phase 2.