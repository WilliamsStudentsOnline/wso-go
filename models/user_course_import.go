package models

// Importing resets every term through the transcript's last; reviewed courses and later terms stay

// A letter the catalog doesn't use (a transcript's "397T") falls back to the bare number
func (m *CourseCatalogModel) TermCourseByCode(subject string, number int, letter string, strm int) (*CourseInSemester, error) {
	var cis CourseInSemester
	err := m.db.
		Preload("CourseCanonical").
		Joins(`JOIN course_listings cl ON cl.course_canonical_id = course_in_semesters.course_canonical_id
  AND cl.deleted_at IS NULL AND cl.subject = ? AND cl.number = ? AND cl.letter IN (?, '')`, subject, number, letter).
		Where("course_in_semesters.strm = ?", strm).
		Order("cl.letter DESC"). // the exact letter before the bare number
		First(&cis).Error
	if err != nil {
		return nil, err
	}
	return &cis, nil
}

type TranscriptImport struct {
	Through  int
	Add      []*CourseInSemester
	Keep     []*UserCourse
	Remove   []*UserCourse // in a covered term (or someday) but not on the transcript
	Reviewed []*UserCourse // would be removed, but reviewed, so they stay
}

func (m *UserCourseModel) ImportTranscript(userID uint, courses []*CourseInSemester, dryRun bool) (*TranscriptImport, error) {
	plan := &TranscriptImport{}
	onTranscript := map[uint]bool{} // course_in_semester IDs
	taken := map[uint]bool{}        // canonical IDs
	for _, cis := range courses {
		onTranscript[cis.ID] = true
		taken[cis.CourseCanonicalID] = true
		if cis.Strm > plan.Through {
			plan.Through = cis.Strm
		}
	}

	tx := m.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.RollbackUnlessCommitted()
	inTx := &UserCourseModel{db: tx}

	var list []*UserCourse
	err := inTx.preloaded().Set("gorm:query_option", "FOR UPDATE").
		Where("user_id = ?", userID).Order("id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	reviewed, err := inTx.ReviewedEntryIDs(userID)
	if err != nil {
		return nil, err
	}
	have := map[uint]bool{}
	for _, uc := range list {
		switch {
		case uc.CourseInSemesterID != nil && onTranscript[*uc.CourseInSemesterID]:
			have[*uc.CourseInSemesterID] = true
			plan.Keep = append(plan.Keep, uc)
		case uc.CourseInSemester != nil && uc.CourseInSemester.Strm > plan.Through:
			// a later term: plans stay
		case uc.CourseInSemesterID == nil && !taken[uc.CourseCanonicalID]:
			// someday, and not taken yet
		case reviewed[uc.ID]:
			plan.Reviewed = append(plan.Reviewed, uc)
		default:
			plan.Remove = append(plan.Remove, uc)
		}
	}
	for _, cis := range courses {
		if !have[cis.ID] {
			have[cis.ID] = true
			plan.Add = append(plan.Add, cis)
		}
	}
	if dryRun {
		return plan, nil
	}

	for _, uc := range plan.Remove {
		if err := tx.Unscoped().Where("user_course_id = ?", uc.ID).Delete(&UserCourseOffering{}).Error; err != nil {
			return nil, err
		}
		if err := tx.Unscoped().Delete(uc).Error; err != nil {
			return nil, err
		}
	}
	for _, cis := range plan.Add {
		if _, err := inTx.addInTx(userID, cis.CourseCanonicalID, cis, nil); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return plan, nil
}
