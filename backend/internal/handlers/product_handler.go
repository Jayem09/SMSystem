package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"smsystem-backend/internal/database"
	"smsystem-backend/internal/models"
	"smsystem-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	LogService *services.LogService
}

func NewProductHandler(logSvc *services.LogService) *ProductHandler {
	return &ProductHandler{LogService: logSvc}
}

type productInput struct {
	Name        string  `json:"name" binding:"required,min=2,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	CostPrice   float64 `json:"cost_price" binding:"min=0"`
	Stock       int     `json:"stock" binding:"min=0"`
	Size        string  `json:"size"`
	ParentID    *uint   `json:"parent_id"`
	ImageURL    string  `json:"image_url"`
	CategoryID  uint    `json:"category_id" binding:"required"`
	BrandID     uint    `json:"brand_id" binding:"required"`

	PCD         string `json:"pcd"`
	OffsetET    string `json:"offset_et"`
	Width       string `json:"width"`
	Bore        string `json:"bore"`
	Finish      string `json:"finish"`
	SpeedRating string `json:"speed_rating"`
	LoadIndex   string `json:"load_index"`
	DOTCode     string `json:"dot_code"`
	PlyRating   string `json:"ply_rating"`

	IsService         bool  `json:"is_service"`
	PointsRequired    int   `json:"points_required"`
	IsReward          bool  `json:"is_reward"`
	PrimarySupplierID *uint `json:"primary_supplier_id"`
}

func (h *ProductHandler) List(c *gin.Context) {
	branchID, _ := GetUintFromContext(c, "branchID")
	userRole, _ := c.Get("userRole")

	// Get filter params
	search := c.Query("search")
	categoryIDStr := c.Query("category_id")
	brandIDStr := c.Query("brand_id")

	// Parse filter IDs
	var categoryID, brandID uint
	if categoryIDStr != "" {
		if parsed, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil {
			categoryID = uint(parsed)
		}
	}
	if brandIDStr != "" {
		if parsed, err := strconv.ParseUint(brandIDStr, 10, 64); err == nil {
			brandID = uint(parsed)
		}
	}

	// For super_admin, allow explicit branch selection via query param
	if userRole == "super_admin" {
		branchQuery := c.Query("branch_id")
		if branchQuery == "ALL" {
			branchID = 0 // Global stock
		} else if branchQuery != "" {
			if parsedBranchID, err := strconv.ParseUint(branchQuery, 10, 64); err == nil {
				if parsedBranchID > 0 {
					branchID = uint(parsedBranchID)
				}
			}
		}
	}

	// Build query with GORM
	query := database.DB.Model(&models.Product{}).Where("deleted_at IS NULL")

	// Apply search filter
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// Apply category filter
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	// Apply brand filter
	if brandID > 0 {
		query = query.Where("brand_id = ?", brandID)
	}

	// Get products
	var products []models.Product
	if err := query.Order("created_at DESC").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	// Build results with stock
	type productResult struct {
		ID               uint    `json:"id"`
		Name             string `json:"name"`
		Description      string `json:"description"`
		Price            float64 `json:"price"`
		CostPrice        float64 `json:"cost_price"`
		BranchStock     int     `json:"branch_stock"`
		Size             string `json:"size"`
		ParentID         *uint  `json:"parent_id"`
		ImageURL        string `json:"image_url"`
		CategoryID     uint   `json:"category_id"`
		BrandID         uint   `json:"brand_id"`
		ReorderLevel    *int    `json:"reorder_level"`
		PrimarySupplierID *uint  `json:"primary_supplier_id"`
		IsService       bool    `json:"is_service"`
		PCD            string `json:"pcd"`
		OffsetET       string `json:"offset_et"`
		Width          string `json:"width"`
		Bore           string `json:"bore"`
		Finish         string `json:"finish"`
		SpeedRating    string `json:"speed_rating"`
		LoadIndex     string `json:"load_index"`
		DOTCode       string `json:"dot_code"`
		PlyRating     string `json:"ply_rating"`
		PointsRequired int    `json:"points_required"`
		IsReward       bool    `json:"is_reward"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
		Category      interface{} `json:"category"`
		Brand         interface{} `json:"brand"`
		Supplier      interface{} `json:"supplier"`
	}

	results := make([]map[string]interface{}, 0, len(products))
	for _, p := range products {
		// Get stock for this product
		var stock int
		if branchID != 0 {
			database.DB.Raw("SELECT COALESCE(SUM(quantity), 0) FROM batches WHERE product_id = ? AND branch_id = ?", p.ID, branchID).Scan(&stock)
		} else {
			database.DB.Raw("SELECT COALESCE(SUM(quantity), 0) FROM batches WHERE product_id = ?", p.ID).Scan(&stock)
		}

		// Get category name
		var catName, brandName, supName string
		if p.CategoryID > 0 {
			var cat models.Category
			if err := database.DB.First(&cat, p.CategoryID).Error; err == nil {
				catName = cat.Name
			}
		}
		if p.BrandID > 0 {
			var brand models.Brand
			if err := database.DB.First(&brand, p.BrandID).Error; err == nil {
				brandName = brand.Name
			}
		}
		if p.PrimarySupplierID != nil && *p.PrimarySupplierID > 0 {
			var sup models.Supplier
			if err := database.DB.First(&sup, *p.PrimarySupplierID).Error; err == nil {
				supName = sup.Name
			}
		}

		r := map[string]interface{}{
			"id": p.ID,
			"name": p.Name,
			"description": p.Description,
			"price": p.Price,
			"cost_price": p.CostPrice,
			"branch_stock": stock,
			"size": p.Size,
			"parent_id": p.ParentID,
			"image_url": p.ImageURL,
			"category_id": p.CategoryID,
			"brand_id": p.BrandID,
			"reorder_level": p.ReorderLevel,
			"primary_supplier_id": p.PrimarySupplierID,
			"is_service": p.IsService,
			"pcd": p.PCD,
			"offset_et": p.OffsetET,
			"width": p.Width,
			"bore": p.Bore,
			"finish": p.Finish,
			"speed_rating": p.SpeedRating,
			"load_index": p.LoadIndex,
			"dot_code": p.DOTCode,
			"ply_rating": p.PlyRating,
			"points_required": p.PointsRequired,
			"is_reward": p.IsReward,
			"created_at": p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at": p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			"category": map[string]interface{}{"name": catName},
			"brand": map[string]interface{}{"name": brandName},
			"supplier": map[string]interface{}{"name": supName},
		}
		results = append(results, r)
	}

	c.JSON(http.StatusOK, gin.H{"products": results})
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	branchID, _ := GetUintFromContext(c, "branchID")

	type ProductWithStock struct {
		models.Product
		Stock int `json:"stock"`
	}

	var queryArgs []interface{}
	var stockSubquery string
	if branchID != 0 {
		stockSubquery = "(SELECT COALESCE(SUM(quantity), 0) FROM batches WHERE batches.product_id = products.id AND batches.branch_id = ?)"
		queryArgs = append(queryArgs, branchID)
	} else {
		stockSubquery = "(SELECT COALESCE(SUM(quantity), 0) FROM batches WHERE batches.product_id = products.id)"
	}

	var productWithStock ProductWithStock
	if err := database.DB.Raw(`
		SELECT p.*, `+stockSubquery+` as stock 
		FROM products p 
		WHERE p.id = ? AND p.deleted_at IS NULL`,
		append(queryArgs, id)...).
		Scan(&productWithStock).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	product = productWithStock.Product
	product.Stock = productWithStock.Stock

	database.DB.Preload("Category").Preload("Brand").First(&product, id)
	c.JSON(http.StatusOK, gin.H{"product": product})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var input productInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	var category models.Category
	if err := database.DB.First(&category, input.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found"})
		return
	}

	var brand models.Brand
	if err := database.DB.First(&brand, input.BrandID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Brand not found"})
		return
	}

	product := models.Product{
		Name:              input.Name,
		Description:       input.Description,
		Price:             input.Price,
		CostPrice:         input.CostPrice,
		Stock:             input.Stock,
		Size:              input.Size,
		ParentID:          input.ParentID,
		ImageURL:          input.ImageURL,
		CategoryID:        input.CategoryID,
		BrandID:           input.BrandID,
		PCD:               input.PCD,
		OffsetET:          input.OffsetET,
		Width:             input.Width,
		Bore:              input.Bore,
		Finish:            input.Finish,
		SpeedRating:       input.SpeedRating,
		LoadIndex:         input.LoadIndex,
		DOTCode:           input.DOTCode,
		PlyRating:         input.PlyRating,
		IsService:         input.IsService,
		PointsRequired:    input.PointsRequired,
		IsReward:          input.IsReward,
		PrimarySupplierID: input.PrimarySupplierID,
	}

	bID, _ := GetUintFromContext(c, "branchID")
	userID, _ := GetUintFromContext(c, "userID")

	if bID == 0 {
		bID = 1
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			return err
		}

		if input.Stock > 0 && !input.IsService {
			warehouse, err := getOrCreateBranchWarehouse(tx, bID)
			if err != nil {
				return err
			}

			batch := models.Batch{
				ProductID:   product.ID,
				WarehouseID: warehouse.ID,
				BranchID:    bID,
				BatchNumber: "INITIAL",
				Quantity:    input.Stock,
			}
			if err := tx.Create(&batch).Error; err != nil {
				return err
			}

			var userIDPtr *uint
			if userID != 0 {
				uid := userID
				userIDPtr = &uid
			}

			movement := models.StockMovement{
				ProductID:   product.ID,
				BatchID:     &batch.ID,
				WarehouseID: warehouse.ID,
				BranchID:    bID,
				UserID:      userIDPtr,
				Type:        models.MovementTypeIn,
				Quantity:    input.Stock,
				Reference:   "Initial Stock upon Creation",
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}
		return syncProductStock(tx, product.ID)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product and initialize stock: " + err.Error()})
		return
	}

	database.DB.Preload("Category").Preload("Brand").
		Select("products.*, (SELECT COALESCE(SUM(quantity), 0) FROM batches WHERE product_id = products.id AND branch_id = ?) as stock", bID).
		First(&product, product.ID)

	h.LogService.Record(userID, "CREATE", "Product", strconv.Itoa(int(product.ID)), fmt.Sprintf("Created product: %s with initial stock %d", product.Name, input.Stock), c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"message": "Product created", "product": product})
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var input productInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	oldPrice := product.Price
	product.Name = input.Name
	product.Description = input.Description
	product.Price = input.Price
	product.CostPrice = input.CostPrice
	product.Size = input.Size
	product.ParentID = input.ParentID
	product.ImageURL = input.ImageURL
	product.CategoryID = input.CategoryID
	product.BrandID = input.BrandID
	product.PCD = input.PCD
	product.OffsetET = input.OffsetET
	product.Width = input.Width
	product.Bore = input.Bore
	product.Finish = input.Finish
	product.SpeedRating = input.SpeedRating
	product.LoadIndex = input.LoadIndex
	product.DOTCode = input.DOTCode
	product.PlyRating = input.PlyRating
	product.IsService = input.IsService
	product.PointsRequired = input.PointsRequired
	product.IsReward = input.IsReward

	bID, _ := GetUintFromContext(c, "branchID")
	userID, _ := GetUintFromContext(c, "userID")

	if bID == 0 {
		bID = 1
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {

		var currentStock int
		if err := tx.Model(&models.Batch{}).
			Where("product_id = ? AND branch_id = ?", product.ID, bID).
			Select("COALESCE(SUM(quantity), 0)").
			Row().Scan(&currentStock); err != nil {
			return err
		}

		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		if !input.IsService && input.Stock != currentStock {
			diff := input.Stock - currentStock

			warehouse, err := getOrCreateBranchWarehouse(tx, bID)
			if err != nil {
				return err
			}

			batch := models.Batch{
				ProductID:   product.ID,
				WarehouseID: warehouse.ID,
				BranchID:    bID,
				BatchNumber: fmt.Sprintf("ADJ-%s", time.Now().Format("20060102")),
				Quantity:    diff,
			}
			if err := tx.Create(&batch).Error; err != nil {
				return err
			}

			var userIDPtr *uint
			if userID != 0 {
				uid := userID
				userIDPtr = &uid
			}

			movement := models.StockMovement{
				ProductID:   product.ID,
				BatchID:     &batch.ID,
				WarehouseID: warehouse.ID,
				BranchID:    bID,
				UserID:      userIDPtr,
				Type:        models.MovementTypeAdjustment,
				Quantity:    diff,
				Reference:   fmt.Sprintf("Direct Edit Sync (From %d to %d)", currentStock, input.Stock),
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}
		return syncProductStock(tx, product.ID)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product and sync stock: " + err.Error()})
		return
	}

	if oldPrice != product.Price {
		h.LogService.Record(userID, "UPDATE_PRICE", "Product", strconv.Itoa(int(product.ID)), fmt.Sprintf("Price changed for %s: P%.2f -> P%.2f", product.Name, oldPrice, product.Price), c.ClientIP())
	} else {
		h.LogService.Record(userID, "UPDATE", "Product", strconv.Itoa(int(product.ID)), fmt.Sprintf("Updated details/stock for %s", product.Name), c.ClientIP())
	}

	database.DB.Preload("Category").Preload("Brand").
		Select("products.*, (SELECT COALESCE(SUM(quantity), 0) FROM batches WHERE product_id = products.id AND branch_id = ?) as stock", bID).
		First(&product, product.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Product updated", "product": product})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if err := database.DB.Delete(&models.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	userID, _ := GetUintFromContext(c, "userID")
	h.LogService.Record(userID, "DELETE", "Product", strconv.Itoa(int(id)), fmt.Sprintf("Deleted product: %s", product.Name), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}
