package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type BoardPollModel struct {
	*BaseModel
}

func NewBoardPollModel(db *gorm.DB, log *zap.SugaredLogger) *BoardPollModel {
	return &BoardPollModel{BaseModel: NewBaseModel(db, log)}
}

func preloadPollOptions(db *gorm.DB) *gorm.DB {
	return db.Order("board_poll_options.position asc")
}

func (m *BoardPollModel) GetPollByThreadID(threadID uint, p *BoardPoll) error {
	return m.DB.Preload("Options", preloadPollOptions).Where("thread_id = ?", threadID).First(p).Error
}

func (m *BoardPollModel) GetVoteOptionID(pollID, userID uint) (*uint, error) {
	var vote BoardPollVote
	err := m.DB.Where("poll_id = ? AND user_id = ?", pollID, userID).First(&vote).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &vote.OptionID, nil
}

func (m *BoardPollModel) CountVotes(pollID uint) (map[uint]int, error) {
	rows, err := m.DB.Model(&BoardPollVote{}).
		Select("option_id, COUNT(*)").
		Where("poll_id = ?", pollID).
		Group("option_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[uint]int{}
	for rows.Next() {
		var optionID uint
		var count int
		if err := rows.Scan(&optionID, &count); err != nil {
			return nil, err
		}
		counts[optionID] = count
	}
	return counts, rows.Err()
}

func (m *BoardPollModel) SetVote(pollID, userID, optionID uint) error {
	now := time.Now()
	return m.DB.Exec(
		"INSERT INTO board_poll_votes (created_at, updated_at, poll_id, user_id, option_id) VALUES (?, ?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE option_id = VALUES(option_id), updated_at = VALUES(updated_at)",
		now, now, pollID, userID, optionID,
	).Error
}
