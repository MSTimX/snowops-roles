package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/MSTimX/Snowops-roles/internal/database"
	"github.com/MSTimX/Snowops-roles/internal/models"
)

type CreateOrganizationRequest struct {
	Name          string `json:"name" binding:"required"`
	Type          string `json:"type" binding:"required"`
	BIN           string `json:"bin"`
	HeadFullName  string `json:"head_full_name"`
	Address       string `json:"address"`
	Phone         string `json:"phone"`
	AdminFullName string `json:"admin_full_name"`
	AdminPhone    string `json:"admin_phone"`
	AdminPassword string `json:"admin_password"`
}

type CreateDriverRequest struct {
	FullName  string `json:"full_name" binding:"required"`
	IIN       string `json:"iin" binding:"required"`
	BirthYear int    `json:"birth_year" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
}

// RegisterRoutes регистрирует HTTP-маршруты для API.
func RegisterRoutes(api *gin.RouterGroup) {
	api.GET("/organizations", ListOrganizations)
	api.POST("/organizations", CreateOrganization)
	api.GET("/organizations/:id", GetOrganization)
	api.PUT("/organizations/:id", UpdateOrganization)
	api.DELETE("/organizations/:id", DeleteOrganization)

	api.GET("/users", FindUser)
	api.GET("/users/:id", GetUser)
	api.PUT("/users/:id", UpdateUser)

	api.GET("/drivers", ListDrivers)
	api.POST("/drivers", CreateDriver)
	api.GET("/drivers/:id", GetDriver)
	api.PUT("/drivers/:id", UpdateDriver)
	api.DELETE("/drivers/:id", DeleteDriver)
}

func ListOrganizations(c *gin.Context) {
	role := c.GetString("currentUserRole")
	currentOrgID := c.GetString("currentOrgID")

	if role == "" || currentOrgID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	currentOrgUUID, err := uuid.Parse(currentOrgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid current organization id"})
		return
	}

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var orgs []models.Organization

	switch role {
	case models.RoleAkimatAdmin:
		if err := database.DB.Where("is_active = ?", true).Find(&orgs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organizations"})
			return
		}
	case models.RoleTooAdmin:
		var currentOrg models.Organization
		if err := database.DB.Where("id = ? AND is_active = ?", currentOrgUUID, true).First(&currentOrg).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization"})
			}
			return
		}

		orgs = append(orgs, currentOrg)

		var contractors []models.Organization
		if err := database.DB.Where("parent_org_id = ? AND type = ? AND is_active = ?", currentOrgUUID, models.OrgTypeContractor, true).Find(&contractors).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch contractor organizations"})
			return
		}

		orgs = append(orgs, contractors...)
	case models.RoleContractorAdmin:
		var currentOrg models.Organization
		if err := database.DB.Where("id = ? AND is_active = ?", currentOrgUUID, true).First(&currentOrg).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization"})
			}
			return
		}
		orgs = append(orgs, currentOrg)
	case models.RoleDriver:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organizations": orgs})
}

func CreateOrganization(c *gin.Context) {
	currentUserID := c.GetString("currentUserID")
	currentUserRole := c.GetString("currentUserRole")
	currentOrgID := c.GetString("currentOrgID")

	if currentUserID == "" || currentUserRole == "" || currentOrgID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !models.CanCreateOrganization(currentUserRole, req.Type) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	currentOrgUUID, err := uuid.Parse(currentOrgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid current organization id"})
		return
	}

	if req.AdminPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin_phone is required"})
		return
	}

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	parentOrgID := currentOrgUUID
	org := models.Organization{
		Type:         req.Type,
		Name:         req.Name,
		BIN:          req.BIN,
		HeadFullName: req.HeadFullName,
		Address:      req.Address,
		Phone:        req.Phone,
		ParentOrgID:  &parentOrgID,
		IsActive:     true,
	}

	if err := tx.Create(&org).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	var adminRole string
	switch req.Type {
	case models.OrgTypeToo:
		adminRole = models.RoleTooAdmin
	case models.OrgTypeContractor:
		adminRole = models.RoleContractorAdmin
	default:
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported organization type"})
		return
	}

	user := models.User{
		Phone:          req.AdminPhone,
		Role:           adminRole,
		OrganizationID: &org.ID,
		IsActive:       true,
	}

	if req.AdminPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		password := string(hashed)
		user.PasswordHash = &password
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin user"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"organization": gin.H{
			"id":           org.ID,
			"name":         org.Name,
			"type":         org.Type,
			"bin":          org.BIN,
			"headFullName": org.HeadFullName,
			"address":      org.Address,
			"phone":        org.Phone,
			"parentOrgID":  org.ParentOrgID,
			"isActive":     org.IsActive,
			"createdAt":    org.CreatedAt,
			"updatedAt":    org.UpdatedAt,
		},
		"admin": gin.H{
			"id":             user.ID,
			"phone":          user.Phone,
			"role":           user.Role,
			"organizationID": user.OrganizationID,
			"isActive":       user.IsActive,
			"createdAt":      user.CreatedAt,
			"updatedAt":      user.UpdatedAt,
		},
	})
}

func GetOrganization(c *gin.Context) {
	role := c.GetString("currentUserRole")
	currentOrgID := c.GetString("currentOrgID")

	if role == "" || currentOrgID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	currentOrgUUID, err := uuid.Parse(currentOrgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid current organization id"})
		return
	}

	targetID := c.Param("id")
	orgUUID, err := uuid.Parse(targetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var org models.Organization
	if err := database.DB.Where("id = ? AND is_active = ?", orgUUID, true).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization"})
		}
		return
	}

	switch role {
	case models.RoleAkimatAdmin:
	case models.RoleTooAdmin:
		if org.ID != currentOrgUUID {
			if org.Type != models.OrgTypeContractor || org.ParentOrgID == nil || *org.ParentOrgID != currentOrgUUID {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
		}
	case models.RoleContractorAdmin:
		if org.ID != currentOrgUUID {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	case models.RoleDriver:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": org})
}

func UpdateOrganization(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}

func DeleteOrganization(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}

func FindUser(c *gin.Context) {
	phone := c.Query("phone")
	login := c.Query("login")

	if phone == "" && login == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone or login required"})
		return
	}

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var user models.User
	q := database.DB.Model(&models.User{}).Where("is_active = ?", true)

	if phone != "" {
		q = q.Where("phone = ?", phone)
	}
	if login != "" {
		q = q.Where("login = ?", login)
	}

	if err := q.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db query failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetUser(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}

func UpdateUser(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}

func ListDrivers(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}

func CreateDriver(c *gin.Context) {
	role := c.GetString("currentUserRole")
	currentOrgID := c.GetString("currentOrgID")

	if role == "" || currentOrgID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if role != models.RoleContractorAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contractorUUID, err := uuid.Parse(currentOrgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid current organization id"})
		return
	}

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	contractorID := contractorUUID
	driver := models.Driver{
		ContractorID: &contractorID,
		FullName:     req.FullName,
		IIN:          req.IIN,
		BirthYear:    req.BirthYear,
		Phone:        req.Phone,
		IsActive:     true,
	}

	if err := tx.Create(&driver).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create driver"})
		return
	}

	driverID := driver.ID
	user := models.User{
		Phone:          req.Phone,
		Role:           models.RoleDriver,
		OrganizationID: &contractorID,
		DriverID:       &driverID,
		IsActive:       true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create driver user"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"driver": driver,
		"user": gin.H{
			"id":             user.ID,
			"phone":          user.Phone,
			"role":           user.Role,
			"organizationID": user.OrganizationID,
			"driverID":       user.DriverID,
			"isActive":       user.IsActive,
			"createdAt":      user.CreatedAt,
			"updatedAt":      user.UpdatedAt,
		},
	})
}

func GetDriver(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}

func UpdateDriver(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}

func DeleteDriver(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}
