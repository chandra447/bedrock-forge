from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional
from mangum import Mangum
import os

# FastAPI app
app = FastAPI(
    title="Product Search API",
    description="API for searching and filtering products",
    version="1.0.0"
)

# Pydantic models
class ProductSearchRequest(BaseModel):
    query: str
    category: Optional[str] = None
    min_price: Optional[float] = None
    max_price: Optional[float] = None
    limit: Optional[int] = 10

class Product(BaseModel):
    id: str
    name: str
    description: str
    price: float
    category: str
    in_stock: bool

class ProductSearchResponse(BaseModel):
    products: List[Product]
    total_count: int
    query: str

# Mock product data
MOCK_PRODUCTS = [
    Product(id="1", name="Wireless Headphones", description="High-quality wireless headphones", price=99.99, category="Electronics", in_stock=True),
    Product(id="2", name="Coffee Maker", description="Automatic drip coffee maker", price=79.99, category="Kitchen", in_stock=True),
    Product(id="3", name="Running Shoes", description="Comfortable running shoes", price=129.99, category="Sports", in_stock=False),
    Product(id="4", name="Laptop Stand", description="Adjustable laptop stand", price=45.99, category="Office", in_stock=True),
    Product(id="5", name="Smartphone", description="Latest model smartphone", price=699.99, category="Electronics", in_stock=True),
]

@app.post("/search", response_model=ProductSearchResponse)
async def search_products(request: ProductSearchRequest):
    """
    Search for products based on query and filters.
    """
    filtered_products = []
    
    for product in MOCK_PRODUCTS:
        # Text search
        if request.query.lower() not in product.name.lower() and request.query.lower() not in product.description.lower():
            continue
            
        # Category filter
        if request.category and product.category.lower() != request.category.lower():
            continue
            
        # Price filters
        if request.min_price and product.price < request.min_price:
            continue
        if request.max_price and product.price > request.max_price:
            continue
            
        filtered_products.append(product)
    
    # Apply limit
    limited_products = filtered_products[:request.limit]
    
    return ProductSearchResponse(
        products=limited_products,
        total_count=len(filtered_products),
        query=request.query
    )

@app.get("/categories")
async def get_categories():
    """
    Get all available product categories.
    """
    categories = list(set(product.category for product in MOCK_PRODUCTS))
    return {"categories": categories}

@app.get("/health")
async def health_check():
    """
    Health check endpoint.
    """
    return {"status": "healthy"}

# Lambda handler
handler = Mangum(app)