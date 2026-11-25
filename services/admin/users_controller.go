package admin

import (
	"net/http"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

const userSearchLimit = 25

type UserBans struct {
	BulletinRead  bool   `json:"bulletinRead"`
	BulletinWrite bool   `json:"bulletinWrite"`
	Factrak       bool   `json:"factrak"`
	Dormtrak      bool   `json:"dormtrak"`
	Reason        string `json:"reason"`
}

type UserAccess struct {
	UserID       uint     `json:"userID"`
	Name         string   `json:"name"`
	UnixID       string   `json:"unixID"`
	Type         string   `json:"type"`
	Admin        bool     `json:"admin"`
	FactrakAdmin bool     `json:"factrakAdmin"`
	Bans         UserBans `json:"bans"`
}

type listUsersResponse struct {
	Users []UserAccess `json:"users"`
}

func newUserAccess(u *models.User, ban *models.BannedUser) UserAccess {
	access := UserAccess{
		UserID:       u.ID,
		Name:         u.Name,
		UnixID:       u.UnixID,
		Type:         u.Type,
		Admin:        u.Admin != nil && *u.Admin,
		FactrakAdmin: u.FactrakAdmin != nil && *u.FactrakAdmin,
	}
	if ban != nil {
		access.Bans = UserBans{
			BulletinRead:  ban.BulletinRead,
			BulletinWrite: ban.BulletinWrite,
			Factrak:       ban.Factrak,
			Dormtrak:      ban.Dormtrak,
			Reason:        ban.Reason,
		}
	}
	return access
}

func (t *Controller) userAccessList(users []*models.User) ([]UserAccess, error) {
	ids := make([]uint, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	bans, err := t.bannedUserModel.GetBannedUsersByUserIDs(ids)
	if err != nil {
		return nil, err
	}
	out := make([]UserAccess, 0, len(users))
	for _, u := range users {
		out = append(out, newUserAccess(u, bans[u.ID]))
	}
	return out, nil
}

// ListUsers godoc
// @Summary List users by search, or admins, moderators and banned users when q is empty
// @ID listAdminUsers
// @Tags admin
// @Produce json
// @Param q query string false "Name or unix ID"
// @Success 200 {object} listUsersResponse
// @Security Bearer
// @Router /admin/users [get]
func (t *Controller) ListUsers(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))

	db := t.db.Model(&models.User{})
	if q != "" {
		like := "%" + q + "%"
		db = db.Where("users.unix_id LIKE ? OR users.name LIKE ?", like, like).
			Order("users.name asc").Limit(userSearchLimit)
	} else {
		db = db.Where(
			"users.admin = ? OR users.factrak_admin = ? OR users.id IN (?)",
			true, true, t.db.Model(&models.BannedUser{}).Select("user_id").QueryExpr(),
		).Order("users.name asc")
	}

	var users []*models.User
	if err := db.Find(&users).Error; err != nil {
		t.RespondError(c, err)
		return
	}

	access, err := t.userAccessList(users)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, listUsersResponse{Users: access})
}

func (t *Controller) loadUserAccess(c *gin.Context) (*models.User, *UserAccess, bool) {
	userID, err := services.GetUIntParam(c, "userID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return nil, nil, false
	}

	user := new(models.User)
	if err := t.userModel.GetUserByID(userID, user); err != nil {
		t.RespondError(c, err)
		return nil, nil, false
	}

	var ban models.BannedUser
	missing, err := t.bannedUserModel.GetBannedUserByID(user.ID, &ban)
	if err != nil {
		t.RespondError(c, err)
		return nil, nil, false
	}
	banPtr := &ban
	if missing {
		banPtr = nil
	}
	access := newUserAccess(user, banPtr)
	return user, &access, true
}

// GetUserAccess godoc
// @Summary Get a user's roles and bans
// @ID getAdminUserAccess
// @Tags admin
// @Produce json
// @Param userID path uint true "User ID"
// @Success 200 {object} UserAccess
// @Security Bearer
// @Router /admin/users/{userID}/access [get]
func (t *Controller) GetUserAccess(c *gin.Context) {
	_, access, ok := t.loadUserAccess(c)
	if !ok {
		return
	}
	t.RespondOK(c, access)
}

type UpdateUserAccessParams struct {
	FactrakAdmin bool     `json:"factrakAdmin"`
	Bans         UserBans `json:"bans"`
}

// UpdateUserAccess godoc
// @Summary Replace a non-admin user's roles and bans
// @Description Takes effect when the user's token is next issued or refreshed
// @ID updateAdminUserAccess
// @Tags admin
// @Accept json
// @Produce json
// @Param userID path uint true "User ID"
// @Param body body UpdateUserAccessParams true "Access"
// @Success 200 {object} UserAccess
// @Security Bearer
// @Router /admin/users/{userID}/access [put]
func (t *Controller) UpdateUserAccess(c *gin.Context) {
	user, access, ok := t.loadUserAccess(c)
	if !ok {
		return
	}

	if access.Admin {
		t.RespondAPIError(c, lib.ErrorAdminUserAccessLocked)
		return
	}

	var params UpdateUserAccessParams
	if err := c.ShouldBind(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	err := t.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", user.ID).
			Update("factrak_admin", params.FactrakAdmin).Error; err != nil {
			return err
		}
		return models.NewBannedUserModel(tx, t.Log).SetBan(user.ID, &models.BannedUser{
			BulletinRead:  params.Bans.BulletinRead,
			BulletinWrite: params.Bans.BulletinWrite,
			Factrak:       params.Bans.Factrak,
			Dormtrak:      params.Bans.Dormtrak,
			Reason:        strings.TrimSpace(params.Bans.Reason),
		})
	})
	if err != nil {
		t.RespondError(c, err)
		return
	}

	_, access, ok = t.loadUserAccess(c)
	if !ok {
		return
	}
	t.RespondOK(c, access)
}
