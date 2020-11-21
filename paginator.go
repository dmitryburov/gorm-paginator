package paginator

import (
	"errors"
	"math"

	"gorm.io/gorm"
)

type Pagination struct {
	TotalRecords int64 `json:"totalRecords"`
	TotalPage    int   `json:"totalPage"`
	Offset       int   `json:"offset"`
	Limit        int   `json:"limit"`
	Page         int   `json:"page"`
	PrevPage     int   `json:"prevPage"`
	NextPage     int   `json:"nextPage"`
}

type Paging struct {
	Page    int      `json:"page"`
	OrderBy []string `json:"orderBy"`
	Limit   int      `json:"limit"`
	ShowSQL bool
}

type Param struct {
	DB     *gorm.DB
	Paging *Paging
}

// Endpoint for pagination
func Pages(p *Param, result interface{}) (paginator *Pagination, err error) {

	var (
		done     = make(chan bool, 1)
		db       = p.DB.Session(&gorm.Session{})
		defPage  = 1
		defLimit = 20
		count    int64
		offset   int
	)

	// get all counts
	go getCounts(db, result, done, &count)

	// if not defined
	if p.Paging == nil {
		p.Paging = &Paging{}
	}

	// debug sql
	if p.Paging.ShowSQL {
		db = db.Debug()
	}
	// limit
	if p.Paging.Limit == 0 {
		p.Paging.Limit = defLimit
	}
	// page
	if p.Paging.Page < 1 {
		p.Paging.Page = defPage
	} else if p.Paging.Page > 1 {
		offset = (p.Paging.Page - 1) * p.Paging.Limit
	}
	// sort
	if len(p.Paging.OrderBy) > 0 {
		for _, o := range p.Paging.OrderBy {
			db = db.Order(o)
		}
	} else {
		str := "id desc"
		p.Paging.OrderBy = append(p.Paging.OrderBy, str)
	}

	// get
	if errGet := db.Limit(p.Paging.Limit).Offset(offset).Find(result).Error; errGet != nil && !errors.Is(errGet, gorm.ErrRecordNotFound) {
		return nil, errGet
	}
	<-done

	// total pages
	total := int(math.Ceil(float64(count) / float64(p.Paging.Limit)))

	// construct pagination
	paginator = &Pagination{
		TotalRecords: count,
		Page:         p.Paging.Page,
		Offset:       offset,
		Limit:        p.Paging.Limit,
		TotalPage:    total,
		PrevPage:     p.Paging.Page,
		NextPage:     p.Paging.Page,
	}

	// prev page
	if p.Paging.Page > 1 {
		paginator.PrevPage = p.Paging.Page - 1
	}
	// next page
	if p.Paging.Page != paginator.TotalPage {
		paginator.NextPage = p.Paging.Page + 1
	}

	return paginator, nil
}

func getCounts(db *gorm.DB, anyType interface{}, done chan bool, count *int64) {
	db.Model(anyType).Count(count)
	done <- true
}

func (p Pagination) IsEmpty() bool {
	return p.TotalRecords <= 0
}
