package goodrich_menu

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
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
	var id uint

	// get menu item id from URL
	id, err = services.GetUIntParam(c, "menuItemID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err) // bad request if error
	}

	// get the menu item from db
	err = t.menuItemModel.GetMenuItem(id, menuItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, menuItem) //respond OK with the menuItem

}

// struct that holds menu item create params
type MenuItemCreateParams struct {
	Title       *string  `json:"title"`
	Description string   `json:"description"`
	Price       *float64 `json:"price"`
	Available   *bool    `json:"available"`
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
		t.RespondError(c, lib.ErrorGoodrichMissingNewMenuItemParams)
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
		err = lib.ErrorGoodrichItemAlreadyExists
		t.RespondError(c, err)
		return

	default: // if some other error, respond with it (error != nil, record not found == false)
		t.RespondError(c, err)
		return
	}


	// construct new item
	var newItem models.MenuItem

	// add fields to new item
	newItem.Title = *lib.StrPtrDefaults(createData.Title, &newItem.Title)
	// trim spaces on description
	newItem.Description = strings.TrimSpace(createData.Description)
	newItem.Price = *lib.FloatPtrDefaults(createData.Price, &newItem.Price)
	newItem.Available = *lib.BoolPtrDefaults(createData.Available, &newItem.Available)


	// create new item in db
	err = t.menuItemModel.CreateMenuItem(&newItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondCreated(c, newItem) // respond created with new item
}

// struct that holds menu item update params
type MenuItemUpdateParams struct {
	Title       *string  `json:"title"`
	Description string   `json:"description"`
	Price       *float64 `json:"price"`
	Available   *bool    `json:"available"`
	// false if item is out of stock
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
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	updateData := MenuItemUpdateParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// get item with menuItemID from db
	var menuItem models.MenuItem
	err = t.menuItemModel.GetMenuItem(menuItemID, &menuItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// update fields
	menuItem.Title = *lib.StrPtrDefaults(updateData.Title, &menuItem.Title)
	// trim spaces on description
	menuItem.Description = strings.TrimSpace(updateData.Description)
	menuItem.Price = *lib.FloatPtrDefaults(updateData.Price, &menuItem.Price)
	menuItem.Available = *lib.BoolPtrDefaults(updateData.Available, &menuItem.Available)

	//update db
	err = t.menuItemModel.UpdateMenuItem(menuItemID, &menuItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	// return updated menu item
	t.RespondOK(c, menuItem)

}