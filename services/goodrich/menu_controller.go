package goodrich

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListMenu godoc
// @Summary List menu
// @Description lists all menu items
// @ID goodrich-list-menu
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param all query bool false "Get all menu items (including unavailable)"
// @Success 200 {array} models.GoodrichMenuItem
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/menu [get]
func (t *Controller) ListMenu(c *gin.Context) {
	params := new(models.GetAllGoodrichMenuItemsOptions)

	err := c.ShouldBindQuery(params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var menu []*models.GoodrichMenuItem
	err = t.menuModel.GetAllGoodrichMenuItems(&menu, params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, menu)
}

// TODO[low]: get menu item

// CreateMenuItemParams is a struct to hold the parameters used to create a menu item.
type CreateMenuItemParams struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Available   bool    `json:"available"`
}

// CreateMenuItem godoc
// @Summary Create menu item
// @Description creates a menu item
// @ID goodrich-create-menu-item
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param createParams body goodrich.CreateMenuItemParams true "Create Menu Item Params"
// @Success 201 {object} models.GoodrichMenuItem
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/menu [post]
func (t *Controller) CreateMenuItem(c *gin.Context) {
	// Bind update params
	createData := CreateMenuItemParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// TODO[high]: validate data

	// Construct new bulletin
	menuitem := models.GoodrichMenuItem{
		Title:       createData.Title,
		Description: createData.Description,
		Price:       createData.Price,
		Available:   createData.Available,
	}

	err = t.menuModel.CreateMenuItem(&menuitem)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, menuitem)
}

// UpdateMenuItemParams is a struct to hold the parameters used to update a menu item.
type UpdateMenuItemParams struct {
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Available   *bool    `json:"available"`
}

// UpdateMenuItem godoc
// @Summary Update menu item
// @Description updates a menu item
// @ID goodrich-update-menu-item
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param updateParams body goodrich.UpdateMenuItemParams true "Update Menu Item Params"
// @Param itemID path uint true "Item ID"
// @Success 200 {object} models.GoodrichMenuItem
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/menu/{itemID} [patch]
func (t *Controller) UpdateMenuItem(c *gin.Context) {
	// Decode parameter
	itemID, err := services.GetUIntParam(c, "itemID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Bind update params
	updateData := UpdateMenuItemParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Do database query to get bulletin
	var menuItem models.GoodrichMenuItem
	err = t.menuModel.GetMenuItemByID(itemID, &menuItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Update fields: this is a bit long and verbose, but I don't want to mess with reflect
	menuItem.Description = *lib.StrPtrDefaults(updateData.Description, &menuItem.Description)
	menuItem.Price = *lib.Float64PtrDefaults(updateData.Price, &menuItem.Price)
	menuItem.Available = *lib.BoolPtrDefaults(updateData.Available, &menuItem.Available)

	// TODO[high]: validate data

	// Update the menu item in the db
	err = t.menuModel.UpdateMenuItem(&menuItem)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Return update menu item
	t.RespondOK(c, menuItem)
}
