package dormtrak

import (
	"fmt"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type DormRoomPhotoInfo struct {
	FileName   string `json:"name"`
	DormRoomID uint   `json:"dormRoomId"`
	ReviewID   uint   `json:"reviewId"`
	Number     uint   `json:"number"`
}

// GetRoomPhotos godoc
// @Summary Get dorm room photos
// @Description gets file names to all photos uploaded to a dorm room
// @ID getRoomPhotos
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param roomID path uint true "Dorm Room ID"
// @Success 200 {array} DormRoomPhotoInfo
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/rooms/{roomID}/photos [get]
func (t *Controller) GetRoomPhotos(c *gin.Context) {
	// Decode roomID.
	roomID, err := services.GetUIntParam(c, "roomID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	photoPaths, err := t.pictureBackend.ListDormRoom(roomID)
	if err != nil {
		if os.IsNotExist(err) {
			t.RespondAPIError(c, lib.ErrorRecordNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	// Assume photos are in format `{reviewID}_{n}.jpg`
	photoInfos := make([]DormRoomPhotoInfo, len(photoPaths))
	for i, photoPath := range photoPaths {
		var fileN, reviewID uint
		_, err := fmt.Sscanf(photoPath, "%d_%d.jpg", &reviewID, &fileN)
		if err != nil {
			continue
		}

		photoInfos[i] = DormRoomPhotoInfo{
			FileName:   photoPath,
			DormRoomID: roomID,
			ReviewID:   reviewID,
			Number:     fileN,
		}
	}

	t.RespondOK(c, photoInfos)
}
