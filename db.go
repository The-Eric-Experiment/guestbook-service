package main

import (
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func arrayFromRows[TResult interface{}](database *Database, rows *sql.Rows) []*TResult {
	defer rows.Close()
	list := make([]*TResult, 0)
	for rows.Next() {
		var output TResult

		database.Scan(rows, &output)

		list = append(list, &output)
	}
	return list
}

type GuestbookRequest struct {
	Message   string `json:"message"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
	Nickname  string `json:"nickname"`
}

type Guestbook struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Created   time.Time `json:"created"`
	Message   string    `json:"message"`
	IP        string    `gorm:"index" json:"ip"`
	UserAgent string    `gorm:"index" json:"userAgent"`
	Nickname  string    `json:"nickname"`
}

type Database struct {
	db *gorm.DB
}

func createId(date time.Time, value string, ip string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(date.String()+value+ip)).String()
}

func (d *Database) Connect(file string) {
	db, err := gorm.Open(sqlite.Open("file:"+file+"?cache=shared&mode=rwc&_journal_mode=WAL"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		// Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		panic("failed to connect database")
	}

	d.db = db
}

func (d *Database) Close() {
	db, err := d.db.DB()
	if err != nil {
		panic("failed to connect database")
	}
	db.Close()
}

func (d *Database) Initialize() {
	// Migrate the schema
	d.db.AutoMigrate(&Guestbook{})
}

func (d *Database) Scan(rows *sql.Rows, dest interface{}) error {
	return d.db.ScanRows(rows, &dest)
}

func (d *Database) AddRawToGuestbook(info Guestbook) {
	d.db.Create(&Guestbook{
		ID:        info.ID,
		Message:   info.Message,
		IP:        info.IP,
		Created:   info.Created,
		UserAgent: info.UserAgent,
		Nickname:  info.Nickname,
	})
}

func (d *Database) AddToGuestbook(info GuestbookRequest) {
	now := time.Now()

	d.db.Create(&Guestbook{
		ID:        createId(now, info.Message, info.IP),
		Message:   info.Message,
		IP:        info.IP,
		Created:   now,
		UserAgent: info.UserAgent,
		Nickname:  info.Nickname,
	})
}

func (d *Database) GetGuestbookCount() int64 {
	var count int64
	d.db.Model(&Guestbook{}).Count(&count)
	return count
}

func (d *Database) GetGuestbook(offset int, limit int) []*Guestbook {
	q := d.db.Model(&Guestbook{}).Order("created DESC").Limit(limit).Offset(offset)

	rows, _ := q.Rows()

	results := arrayFromRows[Guestbook](d, rows)

	return results
}

func (d *Database) GetLastMessage(ua string, ip string) *Guestbook {
	q := d.db.Model(&Guestbook{}).Where(&Guestbook{UserAgent: ua, IP: ip}).Order("created DESC").Limit(1)

	rows, _ := q.Rows()

	results := arrayFromRows[Guestbook](d, rows)
	if len(results) == 0 {
		return nil
	}
	return results[0]
}

func (d *Database) DeleteGuestbook(id string) {
	d.db.Delete(&Guestbook{}, &Guestbook{ID: id})
}

func GetDB(website string) *Database {
	database := Database{}
	path := "data/" + GetDBFilename(website)
	_, err := os.Stat(path)
	newDb := errors.Is(err, os.ErrNotExist)

	database.Connect(path)

	if newDb {
		database.Initialize()
	}

	return &database
}

func GetDBFilename(site string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(site)).String() + ".db"
}
