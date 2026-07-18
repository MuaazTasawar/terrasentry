import os
from dotenv import load_dotenv

load_dotenv()


class Settings:
    PORT: int = int(os.getenv("PORT", "8000"))
    ENV: str = os.getenv("ENV", "development")

    LLM_API_KEY: str = os.getenv("LLM_API_KEY", "")
    LLM_MODEL: str = os.getenv("LLM_MODEL", "claude-sonnet-4-6")

    RISK_SCORE_HIGH_THRESHOLD: int = int(os.getenv("RISK_SCORE_HIGH_THRESHOLD", "75"))
    RISK_SCORE_MEDIUM_THRESHOLD: int = int(os.getenv("RISK_SCORE_MEDIUM_THRESHOLD", "40"))


settings = Settings()