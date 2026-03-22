# SMSystem Documentation
## Sales & Inventory Management System

**Version:** 0.3.0  
**Last Updated:** March 2026

---

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [User Roles & Permissions](#user-roles--permissions)
4. [Dashboard](#dashboard)
5. [Point of Sale (POS)](#point-of-sale-pos)
6. [Products Management](#products-management)
7. [Inventory Management](#inventory-management)
8. [Branches & Transfers](#branches--transfers)
9. [Orders Management](#orders-management)
10. [Customers & CRM](#customers--crm)
11. [Expenses](#expenses)
12. [Suppliers & Purchase Orders](#suppliers--purchase-orders)
13. [Staff Management](#staff-management)
14. [Reports & Logs](#reports--logs)
15. [Settings](#settings)
16. [API Documentation](#api-documentation)
17. [Troubleshooting](#troubleshooting)

---

## Introduction

SMSystem is a comprehensive **Sales & Inventory Management System** designed for multi-branch retail businesses. It provides complete control over inventory, sales, customer relationships, and branch operations.

### Key Features

- **Multi-Branch Support**: Manage multiple branches with isolated inventory
- **Point of Sale**: Fast, intuitive checkout with receipt printing
- **Inventory Tracking**: Real-time stock levels with batch tracking
- **Transfer Management**: Move inventory between branches
- **Customer CRM**: Track customer information and loyalty
- **Purchase Orders**: Manage supplier orders and incoming stock
- **Expense Tracking**: Monitor and categorize business expenses
- **Daily Reports**: Generate sales and inventory reports
- **Role-Based Access**: Granular permissions for different user types

---

## Getting Started

### System Requirements

- **Backend**: Go 1.21+, MySQL 8.0+
- **Frontend**: Node.js 18+, npm/yarn
- **Browser**: Chrome 90+, Firefox 88+, Safari 14+

### Installation

#### 1. Clone the Repository

```bash
git clone https://github.com/Jayem09/SMSystem.git
cd SMSystem
```

#### 2. Backend Setup

```bash
cd backend

# Copy environment file
cp .env.example .env

# Edit .env with your database credentials
# DB_HOST=your-mysql-host
# DB_PORT=3306
# DB_USER=root
# DB_PASSWORD=your-password
# DB_NAME=smsystem

# Install dependencies
go mod download

# Build the server
go build -o smsystem-backend ./cmd/server

# Run the server
./smsystem-backend
```

The server will start on port **8080** by default.

#### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

The frontend will be available at **http://localhost:5173**

### First-Time Setup

1. Navigate to the login page
2. Click "Register" to create your first account
3. Fill in your details (name, email, password)
4. Your account will be assigned to a branch
5. Start using the system!

---

## User Roles & Permissions

SMSystem uses role-based access control (RBAC) with the following roles:

| Role | Description |
|------|-------------|
| **super_admin** | Full system access, can manage all branches |
| **admin** | Branch-level full access, can manage staff and settings |
| **purchasing** | Can manage products, inventory, purchase orders |
| **purchaser** | Limited purchasing access |
| **cashier** | POS access, can process sales |
| **user** | Basic access, can view transfers |

### Role Access Matrix

| Feature | super_admin | admin | purchasing | cashier | user |
|---------|-------------|-------|------------|---------|------|
| Dashboard | ✅ | ✅ | ✅ | ✅ | ❌ |
| POS | ✅ | ✅ | ✅ | ✅ | ❌ |
| Products | ✅ | ✅ | ✅ | ❌ | ❌ |
| Inventory | ✅ | ✅ | ✅ | ❌ | ❌ |
| Orders | ✅ | ✅ | ✅ | ✅ | ❌ |
| Transfers | ✅ | ✅ | ✅ | ✅ | ✅ |
| Customers | ✅ | ✅ | ✅ | ❌ | ❌ |
| Expenses | ✅ | ✅ | ✅ | ❌ | ❌ |
| Suppliers | ✅ | ✅ | ✅ | ❌ | ❌ |
| Staff | ✅ | ✅ | ❌ | ❌ | ❌ |
| Branches | ✅ | ✅ | ❌ | ❌ | ❌ |
| Reports | ✅ | ✅ | ❌ | ❌ | ❌ |
| Settings | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## Dashboard

The Dashboard provides an overview of your branch's performance.

### Access
- **URL**: `/dashboard`
- **Roles**: All authenticated users except `user`

### Features

#### Summary Cards
- **Today's Sales**: Total revenue for today
- **Orders Today**: Number of orders processed
- **Low Stock**: Products below reorder level
- **Pending Transfers**: Transfers awaiting action

#### Quick Actions
- **New Sale**: Opens POS for immediate checkout
- **Receive Stock**: Quick access to stock receiving
- **View Reports**: Access daily reports

#### Recent Activity
- Shows latest orders and activities
- Click to view full details

---

## Point of Sale (POS)

The POS is the primary interface for processing customer transactions.

### Access
- **URL**: `/pos`
- **Roles**: super_admin, admin, purchasing, cashier

### Step-by-Step Checkout

#### 1. Select Products

```
┌─────────────────────────────────────────┐
│ 🔍 Search products...                    │
├─────────────────────────────────────────┤
│ [All] [Tires] [Batteries] [Oils] [...] │
├─────────────────────────────────────────┤
│ ┌─────────────────┬──────┬─────────┐   │
│ │ Product Name    │ Price │ Stock   │   │
│ ├─────────────────┼──────┼─────────┤   │
│ │ ACCELERA Tire   │ ₱2058│   88    │   │
│ │ MOTOLITE Battery│ ₱4402│   45    │   │
│ │ PETRONAS Oil    │ ₱455 │  120    │   │
│ └─────────────────┴──────┴─────────┘   │
└─────────────────────────────────────────┘
```

**Actions:**
- Click a product to add to cart
- Use category filters to narrow down
- Search by product name or code

#### 2. Review Cart

```
┌─────────────────────────────────────────┐
│ 🛒 CART                      Total: ₱0 │
├─────────────────────────────────────────┤
│ [Product]     [Qty]   [Price]  [Remove] │
│ ACCELERA Tire   2    ₱4,116    [X]     │
├─────────────────────────────────────────┤
│ Discount: [____] %                       │
│                                         │
│            [CHECKOUT]                   │
└─────────────────────────────────────────┘
```

#### 3. Checkout Process

Click **CHECKOUT** to open the checkout modal:

```
┌─────────────────────────────────────────┐
│ CHECKOUT                                │
├─────────────────────────────────────────┤
│ Customer: [Select or enter guest...]    │
│                                         │
│ Service Advisor: [Name field]           │
│                                         │
│ Receipt Type: (•) SI  ( ) DR            │
│                                         │
│ TIN: [____________]                     │
│ Business Address: [__________________]   │
│                                         │
│ Withholding Tax: [__] %                  │
│                                         │
│ Payment: (•) Cash  ( ) Card  ( ) GCash │
├─────────────────────────────────────────┤
│ Subtotal:     ₱4,116                    │
│ Discount:     ₱0                        │
│ Tax:          ₱0                        │
│ ─────────────────────────               │
│ TOTAL:        ₱4,116                   │
├─────────────────────────────────────────┤
│        [PRINT RECEIPT]  [COMPLETE]     │
└─────────────────────────────────────────┘
```

**Fields:**
- **Customer**: Select from existing or enter as guest
- **Service Advisor**: Optional field for service jobs
- **Receipt Type**: SI (Sales Invoice) or DR (Delivery Receipt)
- **TIN/Business Address**: For business invoices
- **Payment Method**: Cash, Card, or GCash

#### 4. Complete Transaction

1. Click **COMPLETE** to process the sale
2. Receipt will be printed automatically
3. Stock will be deducted automatically
4. Order is recorded in the system

### Printing Receipts

- **SI Receipt**: Standard sales invoice
- **DR Receipt**: Delivery receipt for item transfers
- Print options available after completion

---

## Products Management

Manage your product catalog with full details including pricing, variants, and specifications.

### Access
- **URL**: `/products`
- **Roles**: super_admin, admin, purchasing, purchaser

### Product Fields

| Field | Description | Required |
|-------|-------------|----------|
| Name | Product name | ✅ |
| Description | Product details | ❌ |
| Category | Product category | ✅ |
| Brand | Product brand | ✅ |
| Price | Selling price | ✅ |
| Cost Price | Purchase cost | ❌ |
| Size | Size/variant (e.g., 225/45 R17) | ❌ |
| Stock | Initial stock level | ❌ |
| Is Service | Mark as service item | ❌ |

### Tire-Specific Fields

| Field | Description |
|-------|-------------|
| PCD | Pitch Circle Diameter |
| Offset ET | Wheel offset |
| Width | Section width |
| Bore | Center bore |
| Finish | Wheel finish |
| Speed Rating | Speed rating (e.g., V, W, Y) |
| Load Index | Load capacity index |
| DOT Code | Department of Transportation code |
| Ply Rating | Ply rating |

### Managing Products

#### Add New Product

1. Click **+ NEW PRODUCT**
2. Fill in the product details form
3. Select category and brand
4. For tires, fill in tire-specific fields
5. Click **SAVE**

#### Edit Product

1. Find the product in the list
2. Click the **Edit** (pencil) icon
3. Modify the details
4. Click **UPDATE**

#### Delete Product

1. Find the product in the list
2. Click the **Delete** (trash) icon
3. Confirm the deletion
4. Product is soft-deleted (can be recovered)

### Product Variants

Products can have variants linked via **Parent ID**:

1. Create a parent product (e.g., "TIRE 225/45 R17")
2. Create variants with size differences
3. Link variants to parent product

---

## Inventory Management

Complete inventory control with batch tracking, stock movements, and warehouse management.

### Access
- **URL**: `/inventory`
- **Roles**: super_admin, admin, purchasing, purchaser

### Tabs

#### 1. Stock Levels

View current stock across all warehouses.

```
┌─────────────────────────────────────────────────┐
│ Stock Levels                                    │
├─────────────────────────────────────────────────┤
│ Warehouse: [All Warehouses ▼] [Export]          │
├─────────────────────────────────────────────────┤
│ Product        │ Warehouse │ Stock │ In Transit │
├────────────────┼───────────┼───────┼───────────┤
│ ACCELERA Tire  │ Main      │  88   │    0      │
│ MOTOLITE Batt  │ Main      │  45   │   10      │
└────────────────┴───────────┴───────┴───────────┘
```

**Features:**
- Filter by warehouse
- View in-transit stock
- Export to Excel

#### 2. Receive Stock (IN)

Record incoming inventory from suppliers.

**Step-by-Step:**

1. Click **+ NEW RECEIVE**
2. Select **Supplier**
3. Select **Warehouse** (your branch warehouse)
4. Add products with:
   - Product name
   - Quantity received
   - Unit cost
5. Enter **Reference** (e.g., PO number, delivery receipt)
6. Click **RECEIVE**

```
┌─────────────────────────────────────────┐
│ Receive Stock                           │
├─────────────────────────────────────────┤
│ Supplier: [Select Supplier ▼]           │
│ Warehouse: [Main Warehouse ▼]           │
│ Reference: [PO-2024-001]               │
├─────────────────────────────────────────┤
│ Products:                               │
│ ┌─────────────────────────────────────┐ │
│ │ [ACCELERA Tire    ] Qty: [10] Cost: │ │
│ │ [MOTOLITE Battery ] Qty: [5]  Cost:  │ │
│ └─────────────────────────────────────┘ │
│ [+ Add Item]                            │
├─────────────────────────────────────────┤
│          [CANCEL]  [RECEIVE STOCK]      │
└─────────────────────────────────────────┘
```

#### 3. Stock Out (OUT)

Record stock removals (damaged, expired, adjustments).

**Step-by-Step:**

1. Click **+ NEW STOCK OUT**
2. Select **Product**
3. Enter **Quantity**
4. Enter **Reference** (reason)
5. Click **SUBMIT**

#### 4. Movement Logs

View complete history of all stock movements.

**Includes:**
- Date and time
- Product name
- Movement type (IN/OUT/ADJUSTMENT)
- Quantity changed
- Reference number
- User who made the change

### Batch Tracking

Each stock movement is tracked in batches:

- **Batch Number**: Auto-generated or manual
- **Expiry Date**: For products with expiration
- **Unit Cost**: Cost per unit at time of receipt
- **追溯**: Full traceability of stock

---

## Branches & Transfers

SMSystem supports multi-branch operations with isolated inventory.

### Access
- **URL**: `/branches` (admin only)
- **URL**: `/transfers` (all users)

### Branch Management (Admin)

#### Add New Branch

1. Go to **Branches**
2. Click **+ NEW BRANCH**
3. Enter branch details:
   - Branch Name
   - Branch Code (unique identifier)
   - Address
   - Phone Number
4. Click **SAVE**

### Transfer System

Transfers allow moving stock between branches.

#### Transfer Flow

```
[Source Branch] ──request──> [Request Sent]
                                    │
                                    ▼
                           [Destination Branch]
                                    │
                                    ▼ (approve/reject)
                             [Approved/Rejected]
                                    │
                                    ▼ (if approved)
                              [In Transit]
                                    │
                                    ▼
                            [Received at Destination]
```

#### Creating a Transfer Request

**Step-by-Step:**

1. Go to **Transfers**
2. Click **+ REQUEST STOCK**
3. Select **Source Branch** (the branch to request from)
4. Add products and quantities
5. Add notes (optional)
6. Click **SUBMIT REQUEST**

```
┌─────────────────────────────────────────┐
│ Request Stock from Another Branch        │
├─────────────────────────────────────────┤
│ Request From: [Select Branch ▼]         │
│                                         │
│ Products:                               │
│ ┌─────────────────────────────────────┐ │
│ │ [Product Name]        Qty: [___] X  │ │
│ └─────────────────────────────────────┘ │
│ [+ Add Product]                         │
│                                         │
│ Notes: [Optional notes for request...]  │
├─────────────────────────────────────────┤
│        [CANCEL]  [SUBMIT REQUEST]      │
└─────────────────────────────────────────┘
```

#### Managing Transfer Requests

**As Requester (Destination Branch):**

| Status | Meaning | Action |
|--------|---------|--------|
| pending | Awaiting source branch approval | Cancel request |
| rejected | Source branch rejected | View reason |
| approved | Source approved, awaiting shipment | Wait for shipping |
| in_transit | Shipped, in delivery | Wait for receipt |
| completed | Received at destination | Done |

**As Source Branch:**

1. View incoming requests in **Incoming** tab
2. Review requested items
3. **Approve**: Accept the transfer
4. **Reject**: Decline with reason
5. **Ship**: Mark as in_transit (this deducts from your stock)
6. Once shipped, await destination receipt confirmation

#### Receiving Transfers

When stock arrives at destination:

1. Go to **Transfers** → **Incoming**
2. Find transfer with **in_transit** status
3. Click **RECEIVE**
4. Confirm receipt
5. Stock is added to destination warehouse

---

## Orders Management

View and manage all sales orders.

### Access
- **URL**: `/orders`
- **Roles**: All users

### Order Statuses

| Status | Description |
|--------|-------------|
| pending | Order created, awaiting payment |
| completed | Payment received, fulfilled |
| cancelled | Order cancelled |
| refunded | Payment refunded |

### Viewing Orders

**Order List:**
- Filter by status
- Search by order ID or customer
- View order details
- Print receipts (SI or DR)
- Process refunds

**Order Details:**
- Customer information
- List of items purchased
- Payment method
- Total amounts
- Timestamp

### Printing Receipts

1. Find the order
2. Click **View** (eye icon)
3. Click **PRINT RECEIPT** or **PRINT DR**

---

## Customers & CRM

Manage customer relationships and track customer purchases.

### Access
- **URL**: `/customers`
- **URL**: `/crm` (admin only)
- **Roles**: All users for basic; Admin for CRM

### Customer Management

#### Add Customer

1. Go to **Customers**
2. Click **+ NEW CUSTOMER**
3. Fill in details:
   - Name
   - Phone Number
   - Email (optional)
   - Address (optional)
   - TIN (for business customers)
4. Click **SAVE**

#### Customer Fields

| Field | Description |
|-------|-------------|
| Name | Customer's full name or business name |
| Phone | Primary contact number |
| Email | Email address |
| Address | Billing/shipping address |
| TIN | Tax Identification Number |
| Total Purchases | Accumulated purchase value |
| Order Count | Number of orders placed |
| Last Visit | Date of last purchase |

### CRM Features (Admin)

**URL**: `/crm`

#### Customer Analytics

- **Top Customers**: By total purchases
- **Recent Activity**: Latest customer interactions
- **Loyalty Tracking**: Monitor repeat customers

---

## Expenses

Track and categorize business expenses.

### Access
- **URL**: `/expenses`
- **Roles**: super_admin, admin, purchasing, purchaser

### Adding Expenses

1. Click **+ NEW EXPENSE**
2. Fill in details:
   - Description
   - Amount
   - Category
   - Date
   - Notes (optional)
3. Attach receipt (optional)
4. Click **SAVE**

### Expense Categories

Default categories:
- Utilities
- Supplies
- Maintenance
- Transportation
- Miscellaneous

### Reports

View expenses by:
- Date range
- Category
- Total amounts

---

## Suppliers & Purchase Orders

Manage suppliers and create purchase orders.

### Access
- **URL**: `/suppliers`
- **URL**: `/purchase-orders`
- **Roles**: super_admin, admin, purchasing, purchaser

### Supplier Management

#### Add Supplier

1. Go to **Suppliers**
2. Click **+ NEW SUPPLIER**
3. Enter details:
   - Company Name
   - Contact Person
   - Phone
   - Email
   - Address
4. Click **SAVE**

### Purchase Orders

Create purchase orders to request stock from suppliers.

#### Creating a PO

1. Go to **Purchase Orders**
2. Click **+ NEW PO**
3. Select **Supplier**
4. Add products with quantities and expected costs
5. Set **Expected Delivery Date**
6. Click **SUBMIT**

**PO Statuses:**

| Status | Description |
|--------|-------------|
| pending | Awaiting supplier confirmation |
| approved | Supplier approved |
| shipped | Supplier shipped |
| received | Stock received at warehouse |
| cancelled | PO cancelled |

#### Receiving Against a PO

1. Open the approved/shipped PO
2. Click **RECEIVE STOCK**
3. Verify quantities match
4. Confirm receipt
5. Stock is added to warehouse

---

## Staff Management

Manage user accounts and permissions.

### Access
- **URL**: `/staff`
- **Roles**: super_admin, admin only

### Adding Staff

1. Go to **Staff**
2. Click **+ NEW STAFF**
3. Fill in details:
   - Name
   - Email
   - Password
   - Role
   - Branch assignment
4. Click **SAVE**

### User Fields

| Field | Description |
|-------|-------------|
| Name | Staff's full name |
| Email | Login email (unique) |
| Password | Initial password |
| Role | User permission level |
| Branch | Assigned branch |
| Status | Active/Inactive |

### Role Assignment

Choose appropriate roles based on job function:
- **Admin**: Branch manager
- **Purchasing**: Inventory manager
- **Cashier**: Sales associate
- **User**: Basic access

---

## Reports & Logs

### Daily Reports

**URL**: `/daily-report`  
**Roles**: admin only

View comprehensive daily reports including:
- Total sales amount
- Number of transactions
- Top-selling products
- Payment method breakdown
- Low stock alerts

### Activity Logs

**URL**: `/logs`  
**Roles**: admin only

Track all system activities:
- User logins
- Product changes
- Inventory movements
- Order processing
- Settings changes

---

## Settings

Configure system-wide settings.

### Access
- **URL**: `/settings`
- **Roles**: admin only

### Available Settings

| Setting | Description |
|---------|-------------|
| Business Name | Your company name |
| Business Address | Default address on receipts |
| TIN | Your Tax Identification Number |
| Service Advisors | List of service advisor names |
| Printer Name | Receipt printer device name |
| Invoice Prefix | Prefix for invoice numbers |
| Receipt Footer | Custom footer message |

---

## API Documentation

SMSystem provides a RESTful API for integrations.

### Base URL

```
http://localhost:8080/api
```

### Authentication

All API requests require a JWT token:

```
Authorization: Bearer <your-token>
```

### Endpoints

#### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/login` | User login |
| POST | `/auth/register` | User registration |
| GET | `/auth/me` | Get current user |

#### Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/products` | List products |
| POST | `/products` | Create product |
| GET | `/products/:id` | Get product |
| PUT | `/products/:id` | Update product |
| DELETE | `/products/:id` | Delete product |

#### Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/orders` | List orders |
| POST | `/orders` | Create order |
| GET | `/orders/:id` | Get order |
| PUT | `/orders/:id` | Update order |

#### Inventory

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/inventory/levels` | Stock levels |
| POST | `/inventory/in` | Receive stock |
| POST | `/inventory/out` | Stock out |
| GET | `/inventory/logs` | Movement logs |

#### Transfers

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/transfers` | List transfers |
| POST | `/transfers` | Create transfer |
| PUT | `/transfers/:id` | Update transfer |

### Example Request

```bash
# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'

# Response
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { "id": 1, "name": "Admin", "email": "admin@example.com" }
}

# Use token for authenticated requests
curl http://localhost:8080/api/products \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

---

## Troubleshooting

### Common Issues

#### Login Problems

**Issue**: Cannot log in
**Solutions**:
1. Verify email and password are correct
2. Check if account is active
3. Clear browser cache and try again
4. Contact admin to reset password

#### Stock Discrepancies

**Issue**: Stock levels don't match
**Solutions**:
1. Check pending transfers (in_transit status)
2. Review movement logs for unauthorized changes
3. Verify all receipts were processed
4. Check for data sync issues

#### Printer Not Working

**Issue**: Receipt won't print
**Solutions**:
1. Check printer is connected and powered on
2. Verify printer name in Settings
3. Check browser's print permissions
4. Try printing to PDF as workaround

#### Slow Performance

**Issue**: System is slow
**Solutions**:
1. Check internet connection
2. Clear browser cache
3. Close unused browser tabs
4. Refresh the page

### Error Messages

| Error | Meaning | Solution |
|-------|---------|----------|
| "Unauthorized" | Invalid or expired token | Log in again |
| "Forbidden" | Insufficient permissions | Contact admin |
| "Not found" | Resource doesn't exist | Check URL |
| "Server error" | Backend issue | Contact support |

### Data Backup

Regular backups are recommended:
1. MySQL database: Use mysqldump
2. Settings: Export from Settings page
3. Store backups in a secure location

---

## Support

For issues or questions:
- **Email**: support@smsystem.com
- **GitHub**: https://github.com/Jayem09/SMSystem
- **Documentation**: https://docs.smsystem.com

---

## Changelog

### v0.3.0 (March 2026)
- Fixed stock-per-branch isolation
- Security patches for SQL injection
- UI improvements (DataTable, StatusBadge)
- UX enhancements

### v0.2.x (Previous versions)
- Initial releases with core features
- POS system
- Inventory management
- Multi-branch support

---

*SMSystem v0.3.0 - Documentation generated March 2026*
