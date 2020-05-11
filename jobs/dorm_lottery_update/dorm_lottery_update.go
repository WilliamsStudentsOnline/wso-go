package dorm_lottery_update

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"github.com/ledongthuc/pdf"
)

type DormBedRow struct {
	Bed      string
	RoomSize string
	RoomType string
}

type exportBeds struct {
	DormBeds   []DormBed `json:"dormBeds"`
	UpdateTime time.Time `json:"updateTime"`
}

type DormBed struct {
	Bed            string `json:"bed"`
	DormRoomNumber string `json:"dormRoomNumber"`
	DormRoomID     *uint  `json:"dormRoomID"`
	DormName       string `json:"dormName"`
	DormID         *uint  `json:"dormID"`
	RoomType       string `json:"roomType"`
	Gender         string `json:"gender"`
}

// "/Users/aidanlloyd-tucker/Downloads/Room_Selection_Rooms_Remaining.pdf"
func ParseLotteryPDF(f io.ReaderAt, size int64) ([]DormBedRow, error) {
	r, err := pdf.NewReader(f, size)
	if err != nil {
		return nil, err
	}

	totalPages := r.NumPage()

	var dormBeds []DormBedRow

	for i := 1; i <= totalPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			var cols []string
			for _, col := range row.Content {
				cols = append(cols, col.S)
			}

			if len(cols) != 3 {
				continue
			}

			if strings.Contains(cols[0], "Room/Bed") {
				continue
			}

			dormBeds = append(dormBeds, DormBedRow{
				Bed:      cols[0],
				RoomSize: cols[1],
				RoomType: cols[2],
			})
		}
	}

	return dormBeds, nil
}

func ParseLotteryDormBedRows(rows []DormBedRow, db *gorm.DB) ([]DormBed, error) {
	var dormBeds []DormBed

	for _, row := range rows {
		roomBedStr := strings.Split(row.Bed, "-")
		dormRoomName := roomBedStr[0]
		bed := roomBedStr[1]

		dormRoomStr := strings.Split(dormRoomName, " ")
		dormName := strings.Join(dormRoomStr[:len(dormRoomStr)-1], " ")
		roomNumber := dormRoomStr[len(dormRoomStr)-1]

		var dorm models.Dorm
		err := db.Model(&models.Dorm{}).Where("dorms.name = ?", dormName).First(&dorm).Error
		if err != nil {
			return nil, err
		}

		var dormRoom models.DormRoom
		err = db.Model(&models.DormRoom{}).Where("dorm_id = ? AND number = ?", dorm.ID, roomNumber).First(&dormRoom).Error
		if err != nil {
			return nil, err
		}

		dormBeds = append(dormBeds, DormBed{
			Bed:            bed,
			DormRoomNumber: roomNumber,
			DormRoomID:     &dormRoom.ID,
			DormName:       dormName,
			DormID:         &dorm.ID,
			RoomType:       row.RoomSize,
			Gender:         row.RoomType,
		})
	}

	return dormBeds, nil
}

func SaveLottery(w io.Writer, dormBeds []DormBed) error {
	lottery := exportBeds{
		DormBeds:   dormBeds,
		UpdateTime: time.Now(),
	}

	return json.NewEncoder(w).Encode(lottery)
}
