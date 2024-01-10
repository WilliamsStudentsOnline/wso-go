package dormtrak

import (
	"fmt"
	"image"
	"net/http"
	"os"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// ListReviews godoc
// @Summary List reviews
// @Description lists all reviews
// @ID dormtrak-list-reviews
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param dormID query uint false "Dorm ID"
// @Param dormRoomID query uint false "Dorm Room ID"
// @Param userID query uint false "User ID"
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param commented query bool false "Restrict to commented reviews"
// @Success 200 {array} models.DormtrakReview
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/reviews [get]
func (t *Controller) ListReviews(c *gin.Context) {
	params := models.GetAllDormtrakReviewsOptions{}

	err := c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// If we specify userID, ensure it is either self or we are admin
	if params.UserID != nil && *params.UserID != services.GetUserID(c) && !auth.HasScope(c, auth.ScopeAdminAll) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	var reviews []*models.DormtrakReview
	err = t.reviewModel.GetAllReviewsWithOptions(&reviews, &params)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromReviews(c, reviews)

	t.RespondOK(c, reviews)
}

// GetReview godoc
// @Summary Get review
// @Description get one review
// @ID dormtrak-get-review
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param reviewID path uint true "Review ID"
// @Success 200 {object} models.DormtrakReview
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/reviews/{reviewID} [get]
func (t *Controller) GetReview(c *gin.Context) {
	// Decode reviewID.
	reviewID, err := services.GetUIntParam(c, "reviewID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var review models.DormtrakReview
	err = t.reviewModel.GetReviewByID(reviewID, &review)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	// Remove user info unless self or admin
	if !auth.CheckIDIsSelf(c, review.UserID) && !auth.HasScope(c, auth.ScopeAdminAll) {
		review.UserID = 0
		review.User = nil
	}

	t.RespondOK(c, review)
}

type ReviewCreateParams struct {
	// Must include this:
	DormRoomID *uint `json:"dormRoomID" binding:"required"`

	// Params:
	Comment          *string `json:"comment"`
	Closet           *string `json:"closet"`
	Flooring         *string `json:"flooring"`
	CommonRoomAccess *bool   `json:"commonRoomAccess"`
	CommonRoomDesc   *string `json:"commonRoomDesc"`
	ThermostatAccess *bool   `json:"thermostatAccess"`
	KeyOrCard        *string `json:"keyOrCard"`
	Noise            *string `json:"noise"`
	BedAdjustable    *bool   `json:"bedAdjustable"`
	PrivateBathroom  *bool   `json:"privateBathroom"`
	BathroomDesc     *string `json:"bathroomDesc"`
	Loudness         *int    `json:"loudness" binding:"omitempty,gte=0,lte=7"`
	Wifi             *int    `json:"wifi" binding:"omitempty,gte=0,lte=7"`
	Location         *int    `json:"location" binding:"omitempty,gte=0,lte=7"`
	Satisfaction     *int    `json:"satisfaction" binding:"omitempty,gte=0,lte=7"`
}

// CreateReview godoc
// @Summary Create review
// @Description create a review
// @ID dormtrak-create-review
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param createParams body dormtrak.ReviewCreateParams true "Create Review Params"
// @Success 201 {object} models.DormtrakReview
// @Failure 1633 {object} services.BaseErrorResponse "user must be a student and could not be found"
// @Failure 1634 {object} services.BaseErrorResponse "user is missing dorm field"
// @Failure 1635 {object} services.BaseErrorResponse "user does not own this dorm room"
// @Failure 1636 {object} services.BaseErrorResponse "review already exists with passed user ID and dorm room ID"
// @Failure 1101 {object} services.BaseErrorResponse "request data validation failed"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/reviews [post]
func (t *Controller) CreateReview(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind create params
	createData := ReviewCreateParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Get the user
	user := new(models.User)
	if err = t.userModel.GetUserByID(userID, user); err != nil {
		// Don't return 404; instead, return student not found
		if gorm.IsRecordNotFoundError(err) {
			err = lib.ErrorReviewStudentNotFound
		}

		t.RespondError(c, err)
		return
	}

	// User must own a dorm
	if user.DormRoomID == nil {
		t.RespondError(c, lib.ErrorReviewMissingDorm)
		return
	}

	// Ensure that you can only review your own dorm.
	// TODO: figure out what to do for reviewing old dorms
	if *createData.DormRoomID != *user.DormRoomID {
		t.RespondError(c, lib.ErrorReviewDormNotOwner)
		return
	}

	dup, err := t.reviewModel.CheckDuplicateReview(user.ID, *createData.DormRoomID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// If duplicate review, return
	if dup {
		t.RespondError(c, lib.ErrorReviewAlreadyExists)
		return
	}

	// Trim spaces off of comment

	// Construct new review
	review := models.DormtrakReview{
		UserID:     user.ID,
		DormRoomID: *createData.DormRoomID,

		// Trim comment of leading/trailing whitespaces
		Comment:          trimIfNotNil(createData.Comment),
		Closet:           trimIfNotNil(createData.Closet),
		Flooring:         trimIfNotNil(createData.Flooring),
		CommonRoomAccess: createData.CommonRoomAccess,
		CommonRoomDesc:   trimIfNotNil(createData.CommonRoomDesc),
		ThermostatAccess: createData.ThermostatAccess,
		KeyOrCard:        trimIfNotNil(createData.KeyOrCard),
		Noise:            trimIfNotNil(createData.Noise),
		BedAdjustable:    createData.BedAdjustable,
		PrivateBathroom:  createData.PrivateBathroom,
		BathroomDesc:     trimIfNotNil(createData.BathroomDesc),
		Loudness:         createData.Loudness,
		Wifi:             createData.Wifi,
		Location:         createData.Location,
		Satisfaction:     createData.Satisfaction,
	}

	err = t.reviewModel.CreateReview(&review)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// New- Brenda (from factrak)
	err = t.userModel.UpdateDormtrakReviewDeficit(user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, review)
}

type ReviewUpdateParams struct {
	Comment          *string `json:"comment"`
	Closet           *string `json:"closet"`
	Flooring         *string `json:"flooring"`
	CommonRoomAccess *bool   `json:"commonRoomAccess"`
	CommonRoomDesc   *string `json:"commonRoomDesc"`
	ThermostatAccess *bool   `json:"thermostatAccess"`
	KeyOrCard        *string `json:"keyOrCard"`
	Noise            *string `json:"noise"`
	BedAdjustable    *bool   `json:"bedAdjustable"`
	PrivateBathroom  *bool   `json:"privateBathroom"`
	BathroomDesc     *string `json:"bathroomDesc"`
	Loudness         *int    `json:"loudness" binding:"omitempty,gte=0,lte=7"`
	Wifi             *int    `json:"wifi" binding:"omitempty,gte=0,lte=7"`
	Location         *int    `json:"location" binding:"omitempty,gte=0,lte=7"`
	Satisfaction     *int    `json:"satisfaction" binding:"omitempty,gte=0,lte=7"`
}

// UpdateReview godoc
// @Summary Update review
// @Description update a review's data
// @ID dormtrak-update-review
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param updateParams body dormtrak.ReviewUpdateParams true "Update Review Params"
// @Param reviewID path uint true "review ID"
// @Success 200 {object} models.DormtrakReview
// @Failure 1101 {object} services.BaseErrorResponse "request data validation failed"
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/reviews/{reviewID} [patch]
func (t *Controller) UpdateReview(c *gin.Context) {
	userID := services.GetUserID(c)

	reviewID, err := services.GetUIntParam(c, "reviewID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Bind update params
	updateData := ReviewUpdateParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Do database query
	var review models.DormtrakReview
	err = t.reviewModel.GetReviewByID(reviewID, &review)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Survey must be owned by user id
	if review.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Update fields: this is a bit long and verbose, but I don't want to mess with reflect

	// Trim comment of leading/trailing whitespaces
	review.Comment = lib.StrPtrDefaults(trimIfNotNil(updateData.Comment), review.Comment)
	review.Closet = lib.StrPtrDefaults(trimIfNotNil(updateData.Closet), review.Closet)
	review.Flooring = lib.StrPtrDefaults(trimIfNotNil(updateData.Flooring), review.Flooring)
	review.CommonRoomAccess = lib.BoolPtrDefaults(updateData.CommonRoomAccess, review.CommonRoomAccess)
	review.CommonRoomDesc = lib.StrPtrDefaults(trimIfNotNil(updateData.CommonRoomDesc), review.CommonRoomDesc)
	review.ThermostatAccess = lib.BoolPtrDefaults(updateData.ThermostatAccess, review.ThermostatAccess)
	review.KeyOrCard = lib.StrPtrDefaults(trimIfNotNil(updateData.KeyOrCard), review.KeyOrCard)
	review.Noise = lib.StrPtrDefaults(trimIfNotNil(updateData.Noise), review.Noise)
	review.BedAdjustable = lib.BoolPtrDefaults(updateData.BedAdjustable, review.BedAdjustable)
	review.PrivateBathroom = lib.BoolPtrDefaults(updateData.PrivateBathroom, review.PrivateBathroom)
	review.BathroomDesc = lib.StrPtrDefaults(trimIfNotNil(updateData.BathroomDesc), review.BathroomDesc)
	review.Loudness = lib.IntPtrDefaults(updateData.Loudness, review.Loudness)
	review.Wifi = lib.IntPtrDefaults(updateData.Wifi, review.Wifi)
	review.Location = lib.IntPtrDefaults(updateData.Location, review.Location)
	review.Satisfaction = lib.IntPtrDefaults(updateData.Satisfaction, review.Satisfaction)

	// Do db update
	err = t.reviewModel.UpdateReview(&review)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// We know user is owner, so don't need to delete user fields
	t.RespondOK(c, review)
}

// DeleteReview godoc
// @Summary Delete review
// @Description delete a review
// @ID dormtrak-delete-review
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param reviewID path uint true "Review ID"
// @Success 200 {object} models.DormtrakReview
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/reviews/{reviewID} [delete]
func (t *Controller) DeleteReview(c *gin.Context) {
	userID := services.GetUserID(c)

	reviewID, err := services.GetUIntParam(c, "reviewID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query
	var review models.DormtrakReview
	err = t.reviewModel.GetReviewByID(reviewID, &review)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Survey must be owned by user id or admin
	if review.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Do db delete
	err = t.reviewModel.DeleteReview(&review)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	//New - Brenda
	user := new(models.User)
	user.ID = review.UserID
	err = t.userModel.UpdateFactrakSurveyDeficit(user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// We know user is owner, so don't need to delete user fields
	t.RespondOK(c, review)
}

// UploadDormRoomPhoto godoc
// @Summary Upload a dorm room photo by review and dorm room
// @Description upload a dorm room review's photo. You may only upload rooms you have reviewed.
// @ID upload-dorm-room-photo
// @Tags dormtrak
// @Accept  multipart/form-data
// @Produce  json
// @Param reviewID path uint true "Review ID"
// @Param file formData file true "Dorm Room Photo"
// @Success 200
// @Failure 1160 {object} services.BaseErrorResponse "unable to save picture"
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/reviews/{reviewID}/photo [put]
func (t *Controller) UploadDormRoomPhoto(c *gin.Context) {
	userID := services.GetUserID(c)

	reviewID, err := services.GetUIntParam(c, "reviewID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query
	var review models.DormtrakReview
	err = t.reviewModel.GetReviewByID(reviewID, &review)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Survey must be owned by user id
	if review.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	formFile, err := c.FormFile("file")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	file, err := formFile.Open()
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	imgScaled := imaging.Fit(img, 600, 600, imaging.Lanczos)
	err = t.pictureBackend.SaveDormRoom(review.DormRoomID, review.ID, imgScaled)
	if err != nil {
		if pictures.IsErrorMaxDormtrakPhotos(err) {
			c.Error(err)
			t.RespondError(c, lib.ErrorDormtrakTooManyPhotos)
			return
		}

		c.Error(err)
		t.RespondError(c, lib.ErrorUnableToSavePicture)
		return
	}

	t.RespondOK(c, nil)
}

// GetReviewPhotos godoc
// @Summary Get dorm room photos by review
// @Description gets file names to all photos uploaded to a dorm room by review
// @ID dormtrak-get-review-photos
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param reviewID path uint true "Review ID"
// @Success 200 {array} DormRoomPhotoInfo
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/reviews/{reviewID}/photos [get]
func (t *Controller) GetReviewPhotos(c *gin.Context) {
	reviewID, err := services.GetUIntParam(c, "reviewID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query
	var review models.DormtrakReview
	err = t.reviewModel.GetReviewByID(reviewID, &review)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	photoPaths, err := t.pictureBackend.ListDormRoom(review.DormRoomID)
	if err != nil {
		if os.IsNotExist(err) {
			t.RespondAPIError(c, lib.ErrorRecordNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	// Assume photos are in format `{reviewID}_{n}.jpg`
	var photoInfos []DormRoomPhotoInfo
	for _, photoPath := range photoPaths {
		var fileN, parsedReviewID uint
		_, err := fmt.Sscanf(photoPath, "%d_%d.jpg", &parsedReviewID, &fileN)
		if err != nil {
			continue
		}

		if parsedReviewID != review.ID {
			continue
		}

		photoInfos = append(photoInfos, DormRoomPhotoInfo{
			FileName:   photoPath,
			DormRoomID: review.DormRoomID,
			ReviewID:   review.ID,
			Number:     fileN,
		})
	}

	t.RespondOK(c, photoInfos)
}

func trimIfNotNil(str *string) *string {
	if str != nil {
		trimmed := strings.TrimSpace(*str)
		return &trimmed
	}
	return str
}
