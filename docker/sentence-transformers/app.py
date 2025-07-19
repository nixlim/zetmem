#!/usr/bin/env python3
"""
Simple FastAPI service for sentence-transformers embeddings
"""

import os
import logging
from typing import List, Dict, Any
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import numpy as np
from sentence_transformers import SentenceTransformer

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global model instance
model = None

class EmbeddingRequest(BaseModel):
    sentences: List[str]
    model: str = None

class EmbeddingResponse(BaseModel):
    embeddings: List[List[float]]

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    global model
    model_name = os.getenv("MODEL_NAME", "all-MiniLM-L6-v2")
    logger.info(f"Loading model: {model_name}")
    
    try:
        model = SentenceTransformer(model_name)
        logger.info(f"Model loaded successfully: {model_name}")
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        raise
    
    yield
    
    # Shutdown
    logger.info("Shutting down embedding service")

app = FastAPI(
    title="Sentence Transformers Embedding Service",
    description="Simple embedding service using sentence-transformers",
    version="1.0.0",
    lifespan=lifespan
)

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    return {"status": "healthy", "model_loaded": True}

@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": "sentence-transformers-embedding",
        "version": "1.0.0",
        "model": os.getenv("MODEL_NAME", "all-MiniLM-L6-v2"),
        "status": "running"
    }

@app.post("/embeddings", response_model=EmbeddingResponse)
async def generate_embeddings(request: EmbeddingRequest):
    """Generate embeddings for the given sentences"""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    
    if not request.sentences:
        raise HTTPException(status_code=400, detail="No sentences provided")
    
    if len(request.sentences) > 100:
        raise HTTPException(status_code=400, detail="Too many sentences (max 100)")
    
    try:
        logger.info(f"Generating embeddings for {len(request.sentences)} sentences")
        
        # Generate embeddings
        embeddings = model.encode(
            request.sentences,
            convert_to_numpy=True,
            show_progress_bar=False
        )
        
        # Convert to list of lists
        embeddings_list = embeddings.tolist()
        
        logger.info(f"Generated embeddings with shape: {embeddings.shape}")
        
        return EmbeddingResponse(embeddings=embeddings_list)
        
    except Exception as e:
        logger.error(f"Error generating embeddings: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to generate embeddings: {str(e)}")

@app.get("/model/info")
async def model_info():
    """Get model information"""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    
    return {
        "model_name": os.getenv("MODEL_NAME", "all-MiniLM-L6-v2"),
        "max_seq_length": getattr(model, 'max_seq_length', 'unknown'),
        "embedding_dimension": model.get_sentence_embedding_dimension(),
        "device": str(model.device) if hasattr(model, 'device') else 'unknown'
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
