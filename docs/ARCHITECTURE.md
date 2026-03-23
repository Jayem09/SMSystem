# SMSystem Architecture

System Overview

flowchart TB
    FE["React + TypeScript Frontend"] --> AX["Axios + Tauri HTTP"]
    AX --> SERVER["Go Gin Backend :8080"]
    SERVER --> HANDLERS["Handlers (Auth, Order, Product, Customer, etc.)"]
    HANDLERS --> MODELS["Models (GORM)"]
    MODELS -> DB["MySQL :3306"]

Frontend Architecture

flowchart LR
    PAGES["Pages: Login, Dashboard, POS, Orders, Products, Customers, Inventory, Analytics..."] --> COMPONENTS["Components: Layout, Sidebar, Modal, Table..."]
    COMPONENTS --> CONTEXT["AuthContext"]
    CONTEXT --> API["Axios Instance"]
    API --> BACKEND["Go Backend :8080"]

Backend Request Flow

sequenceDiagram
    participant C as Tauri App
    participant R as Gin Router
    participant M as Auth Middleware
    participant H as Handler
    participant S as Service
    participant D as MySQL
    
    C->>R: HTTP Request
    R->>M: Verify JWT
    alt Token Valid
        M->>H: Process Request
        H->>S: Business Logic
        S->D: SQL Query
        D-->>S: Result
        S-->>H: Data
        H-->>R: JSON Response
        R-->>C: 200 OK
    else Token Invalid
        M-->>C: 401 Unauthorized
    end

API Endpoints

flowchart TB
    PUBLIC["Public: /api/health, /api/auth/login, /api/auth/register"]
    PROTECTED["Protected: /api/dashboard, /api/analytics, /api/search"]
    READONLY["Read-Only: GET /api/categories, /api/products, /api/orders..."]
    ADMIN["Admin: POST/PUT/DELETE /api/products, /api/expenses..."]
    SUPERADMIN["Super Admin: POST/PUT /api/branches"]
    
    PUBLIC --> PROTECTED
    PROTECTED --> READONLY
    READONLY --> ADMIN
    ADMIN --> SUPERADMIN

Authentication

flowchart LR
    LOGIN["Login"] --> VALID["Validate Credentials"]
    VALID -->|Valid| JWT["Generate JWT"]
    JWT --> RESP["Return {token, user}"]
    VALID -->|Invalid| ERR["Return Error"]
    
    REQUEST["Request"] --> CHECK["Check Bearer Token"]
    CHECK -->|Valid| VERIFY["Verify JWT"]
    VERIFY --> ROLE["Check Role"]
    ROLE -->|OK| HANDLER["Pass to Handler"]
    ROLE -->|Fail| 403["403 Forbidden"]
    CHECK -->|No Token| 401["401 Unauthorized"]

Inventory Flow

flowchart LR
    SI[Stock In] --> MOV[Stock Movements]
    SO[Stock Out] --> MOV
    ADJ[Adjustment] --> MOV
    TRANS[Transfer] --> MOV
    MOV --> BATCH[Product Batches]
    BATCH --> WH[Warehouses]
    BATCH --> LOW[Low Stock Alerts]
    PO[Create PO] --> REC[Receive PO]
    LOW --> AUTO[Auto-Generate Draft POs]
    AUTO --> PO

Analytics Engine

flowchart TB
    Q[Natural Language Query] --> P[Parse Query]
    P --> PAT[Pattern Matching]
    PAT --> SQL[Generate SQL]
    SQL --> EX[Execute Query]
    
    EX --> REV[Revenue]
    EX --> ORD[Orders]
    EX --> PRO[Products]
    EX --> CUST[Customers]
    EX --> EXP[Expenses]
    
    REV --> CHART[Visualize]
    ORD --> CHART
    PRO --> CHART
    CUST --> CHART
    EXP --> CHART

Deployment

flowchart TB
    GITHUB[GitHub Releases] --> MAC_ARM[macOS ARM64]
    GITHUB --> MAC_X86[macOS x86_64]
    GITHUB --> WIN[Windows]
    
    SERVER[Ubuntu VPS :8080] --> BACKEND[Go Backend]
    SERVER --> STATIC[Static Files]
    
    BACKEND --> MYSQL[(MySQL :3306)]

Technology Stack

| Layer | Technology |
|-------|------------|
| Frontend | React 18, TypeScript, TailwindCSS |
| Desktop | Tauri 2.x |
| Backend | Go + Gin |
| Database | MySQL |
| Auth | JWT |
| ORM | GORM |
| CI/CD | GitHub Actions |
