package goodrich_menu

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
)

// ListMenuItems godoc
// @Summary Gets a list of all menu items
// @Description
// @ID
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param includeUnavailable query boolean false "Include unavailable menu items"
// @Success 200 {array} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/menu [get]
func (t *Controller) ListMenuItems(c *gin.Context) {
	var menuItems *[]*models.MenuItem
	var err error

	query := c.Query("includeUnavailable") // get the query? whats the key here?
	convertedQuery, err := strconv.ParseBool(query)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	err = t.menuItemModel.ListMenuItems(convertedQuery, menuItems)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// respond to API call with 200 OK and the data
	t.RespondOK(c, menuItems)
}

// GetMenuItem godoc
// @Summary Gets a menu item
// @Description
// @ID
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Success 200 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/menu/{menuItemID} [get]
func (t *Controller) GetMenuItem(c *gin.Context) {
	var menuItem *models.MenuItem
	var err error

	id := c.Param("menuItemID") // c.Param returns string
	convertedID64, err := strconv.ParseUint(id,10, 64) // convert to uint64
	if err != nil {
		t.RespondError(c, err)
		return
	}
	convertedID := uint(convertedID64) // convert uint64 to uint
	err = t.menuItemModel.GetMenuItem(convertedID, menuItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, menuItem) //respond OK with the menuItem

}

type MenuItemCreateParams struct {
	Title       *string  `json: "title"`
	Description string   `json: "description"`
	Price       *float64 `json: "price"`
	Available   *bool    `json: "available"`
	// false if item is out of stock
}

// CreateMenuItem godoc
// @Summary Creates a menu item
// @Description
// @ID
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param createParams body menu.MenuItemCreateParams true "Create Menu Item Params"
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Failure 2101 {object} lib.APIError
// @Failure 2111 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/menu [post]
func (t *Controller) CreateMenuItem(c *gin.Context) {
	var err error

	createData := MenuItemCreateParams{}
	err = c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// check if title, price and availability are empty, description is optional?
	if createData.Title == nil || (createData.Price == nil || createData.Available == nil) {
		t.RespondError(c, lib.ErrorMissingNewMenuItemParams) // need to write this one into lib
		return
	}

	// check if item that is going to be added already exists
	item := new(models.MenuItem)
	db := t.menuItemModel.DB
	err = db.First(item, createData.Title).Error
	switch {
	case gorm.IsRecordNotFoundError(err) == true:
		// do nothing, item does not already exist

	case err == nil: // if no error, then item already exists
		err = lib.ErrorItemAlreadyExists // TODO create this custom error
		t.RespondError(c, err)
		return

	default: // if some other error, respond with it (error != nil, record not found == false)
		t.RespondError(c, err)
		return
	}

	// dereference the pointers
	title := *createData.Title
	price := *createData.Price
	available := *createData.Available

	// construct new item
	var newItem = models.MenuItem{
		Title:       title,
		Description: strings.TrimSpace(createData.Description),
		Price:       price,
		Available:   available,
	}

	err = t.menuItemModel.CreateMenuItem(&newItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondCreated(c, newItem) // respond created with new item
}

// UpdateMenuItem godoc
// @Summary Creates a menu item
// @Description
// @ID
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param updateParams body menu.MenuItemUpdateParams true "Update Menu Item Params"
// @Success 200 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/menu/{menuItemID} [patch]
func (t *Controller) UpdateMenuItem(c *gin.Context) {
	// get menu item ID from URL
	menuItemID, err := services.GetUIntParam(c, "menuItemID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	updateData := MenuItemCreateParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var menuItem models.MenuItem
	err = t.menuItemModel.GetMenuItem(menuItemID, &menuItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// dereference pointers
	title := *updateData.Title
	price := *updateData.Price
	available := *updateData.Available

	// update fields
	menuItem.Title = title
	menuItem.Description = strings.TrimSpace(updateData.Description)
	menuItem.Price = price
	menuItem.Available = available

	//update db
	err = t.menuItemModel.UpdateMenuItem(menuItemID, &menuItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, menuItem)

}