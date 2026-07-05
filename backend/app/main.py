from fastapi import FastAPI, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from pydantic import BaseModel
from typing import Optional, List
from contextlib import asynccontextmanager
import os
import uuid
import time
import logging
from datetime import datetime

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class AppState:
    resumes_db: dict = {}
    request_count: int = 0

app_state = AppState()

@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("ResumeTake API starting...")
    yield
    logger.info("ResumeTake API shutting down...")

app = FastAPI(
    title="ResumeTake API",
    description="AI简历优化工具 - 智能简历生成与ATS优化",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/api/docs",
    redoc_url="/api/redoc",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    start_time = time.perf_counter()
    response = await call_next(request)
    process_time = time.perf_counter() - start_time
    response.headers["X-Process-Time"] = f"{process_time:.4f}"
    app_state.request_count += 1
    return response

@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    return JSONResponse(
        status_code=422,
        content={"error": "VALIDATION_ERROR", "message": "请求参数验证失败", "details": exc.errors()},
    )

@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException):
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": "HTTP_ERROR", "message": exc.detail},
    )

@app.exception_handler(Exception)
async def general_exception_handler(request: Request, exc: Exception):
    logger.error(f"Unhandled exception: {exc}")
    return JSONResponse(
        status_code=500,
        content={"error": "INTERNAL_ERROR", "message": "服务器内部错误"},
    )

class Resume(BaseModel):
    id: Optional[str] = None
    user_id: Optional[str] = None
    title: str
    content: dict
    target_job: Optional[str] = None
    job_description: Optional[str] = None
    optimized_content: Optional[dict] = None
    ats_score: Optional[float] = None
    keywords: Optional[List[str]] = None
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

class ResumeOptimizeRequest(BaseModel):
    resume_content: dict
    target_job: Optional[str] = None
    job_description: Optional[str] = None

@app.get("/api/health")
async def health_check():
    return {
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "requests_served": app_state.request_count,
    }

@app.post("/api/v1/resumes", status_code=201)
async def create_resume(resume: Resume):
    resume_id = str(uuid.uuid4())
    resume.id = resume_id
    resume.created_at = datetime.now().isoformat()
    resume.updated_at = datetime.now().isoformat()
    app_state.resumes_db[resume_id] = resume.dict()
    return {"success": True, "data": resume}

@app.get("/api/v1/resumes/{resume_id}")
async def get_resume(resume_id: str):
    if resume_id not in app_state.resumes_db:
        raise HTTPException(status_code=404, detail="简历不存在")
    return {"success": True, "data": app_state.resumes_db[resume_id]}

@app.put("/api/v1/resumes/{resume_id}")
async def update_resume(resume_id: str, resume: Resume):
    if resume_id not in app_state.resumes_db:
        raise HTTPException(status_code=404, detail="简历不存在")
    resume.id = resume_id
    resume.updated_at = datetime.now().isoformat()
    app_state.resumes_db[resume_id] = resume.dict()
    return {"success": True, "data": resume}

@app.delete("/api/v1/resumes/{resume_id}")
async def delete_resume(resume_id: str):
    if resume_id not in app_state.resumes_db:
        raise HTTPException(status_code=404, detail="简历不存在")
    del app_state.resumes_db[resume_id]
    return {"success": True, "message": "删除成功"}

@app.post("/api/v1/optimize")
async def optimize_resume(request: ResumeOptimizeRequest):
    optimized_content = {
        "summary": "资深专业人士，拥有丰富的行业经验和卓越的业绩记录。擅长团队协作与项目管理，具备出色的沟通能力和问题解决能力。",
        "experience": [
            {
                "company": "示例科技有限公司",
                "position": "高级产品经理",
                "duration": "2020-至今",
                "highlights": [
                    "领导5人团队完成核心产品开发，用户增长率达到150%",
                    "优化产品流程，将交付周期缩短40%",
                    "建立数据分析体系，驱动产品迭代决策"
                ]
            }
        ],
        "skills": ["项目管理", "数据分析", "AI工具应用", "团队协作", "产品设计"],
        "education": [
            {
                "school": "知名大学",
                "degree": "硕士学位",
                "major": "计算机科学与技术"
            }
        ]
    }

    ats_score = 87.3
    keywords = ["项目管理", "数据分析", "AI应用", "团队协作", "产品优化", "流程改进"]

    suggestions = [
        "建议添加更多量化成果，如具体提升百分比",
        "可以突出AI工具使用经验，提升竞争力",
        "建议在个人简介中加入核心关键词",
        "工作经历建议按STAR法则优化描述"
    ]

    return {
        "success": True,
        "data": {
            "optimized_content": optimized_content,
            "ats_score": ats_score,
            "keywords": keywords,
            "suggestions": suggestions,
        }
    }

@app.get("/api/v1/templates")
async def get_templates():
    templates = [
        {"id": "professional", "name": "专业商务", "description": "适合传统行业和商务岗位", "category": "商务"},
        {"id": "modern", "name": "现代简约", "description": "适合互联网和科技行业", "category": "科技"},
        {"id": "creative", "name": "创意设计", "description": "适合设计和创意岗位", "category": "创意"},
        {"id": "academic", "name": "学术科研", "description": "适合教育和研究岗位", "category": "学术"},
        {"id": "executive", "name": "高管专用", "description": "适合高级管理岗位", "category": "管理"},
        {"id": "minimal", "name": "极简风格", "description": "简洁大方，通用性强", "category": "通用"},
    ]
    return {"success": True, "data": templates}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
