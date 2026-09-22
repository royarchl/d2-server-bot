package database

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"

	"github.com/bwmarrin/discordgo"
)

// Globally accessible
var Db *sql.DB

const (
	StatusCreator    = 0
	StatusRaidLeader = 1
	StatusActive     = 2
	StatusStandby    = 3
)

func EstablishDatabaseConnection() error {
	dbFile := os.Getenv("DATABASE_FILE_PATH")
	var err error

	Db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if err = Db.Ping(); err != nil {
		return err
	}

	return nil
}

func RegisterServerMembers(members []*discordgo.Member) error {
	tx, err := Db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO USER (id, name)
		VALUES (?, ?)
		ON CONFLICT(id) DO NOTHING
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, member := range members {
		if member.User.Bot {
			continue
		}

		res, err := stmt.Exec(member.User.ID, member.User.Username)
		if err != nil {
			log.Fatal(err)
		}

		// Check if member was actually added (non-duplicate)
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}

		// New User
		if rowsAffected != 0 {
			// Do logic of adding them to other database tables
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// A message of the roster should be sent by the function that calls this.
func RosterCreate(raid_name string, author_id string) (int64, error) {
	const ACTIVE int = 1
	// The ID auto-increments and does not need to be included.
	resp, err := Db.Exec(`
		INSERT INTO ROSTER (raid_name, author_id, is_active)
		VALUES (?, ?, ?)`,
		raid_name, author_id, ACTIVE)

	if err != nil {
		return 0, err
	}

	roster_id, err := resp.LastInsertId()
	if err != nil {
		return 0, err
	}

	const POSITION_FIRST int = 0
	const RAID_LEADER int = 1
	resp, err = Db.Exec(`INSERT INTO ROSTER_PARTICIPANTS VALUES (?, ?, ?, ?)`,
		roster_id, author_id, POSITION_FIRST, RAID_LEADER)

	// Needs error check

	return roster_id, nil
}

func RosterUpdateMessageId(message_id string, roster_id int64) error {
	log.Printf("Roster ID: %d\n", roster_id)

	resp, err := Db.Exec(`UPDATE ROSTER SET message_id = (?) WHERE id = (?)`,
		message_id, roster_id)

	// return err

	rows, err := resp.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		log.Println("No lines updated.")
	} else {
		log.Println("Message id successfully updated.")
	}

	return nil
}

func RosterArchive(roster_id int64) error {
	// Potential errors:
	// - roster already archived
	// - caller not roster author

	const INACTIVE int = 0
	_, err := Db.Exec(`UPDATE ROSTER SET is_active = (?) WHERE id = (?)`,
		INACTIVE, roster_id)

	return err
}

func RosterUserJoin(roster_id int64, user_id string, position float64) error {
	// Potential errors:
	// - user already exists in roster
	// - roster is archived
	return nil
}

func RosterUserLeave(roster_id int64, user_id string) error {
	// Potential errors:
	// - user is not in roster
	// - roster is archived
	// - author is trying to leave with no other players in roster
	return nil
}

func RosterUpdateRaidLeader(roster_id int64, previous_leader_id string,
	new_leader_id string) error {

	// Check if current raid leader == user making request.

	// Check if new raid leader exists in this roster.

	return nil
}

func RosterReserveUser(roster_id int64, user_id string) error {
	// Potential errors:
	// - user does not exist in roster
	// - roster is archived
	// - caller is not roster author
	return nil
}

func RosterQueryIdFromUser(raid_name string, author_id string) (int64, error) {
	var rosterId int64
	const ACTIVE int = 1

	err := Db.QueryRow(`
		SELECT id
		FROM ROSTER
		WHERE raid_name = ? AND author_id = ? AND is_active = ?`,
		raid_name, author_id, ACTIVE,
	).Scan(&rosterId)

	if err != nil {
		return 0, err
	}

	return rosterId, nil
}

func RosterQueryIdFromMessage(message_id string) (int64, error) {
	return 0, nil
}

func RosterQueryUserIds(roster_id int64) []string {
	return nil
}

func MarkUserRaidCompletion(user_id string, raid_name string) error {
	const COMPLETED int = 1
	_, err := Db.Exec(`
		INSERT INTO USER_RAID_COMPLETIONS (user_id, raid_id, has_completed)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id, raid_name)
		DO UPDATE SET has_completed = excluded.has_completed`,
		user_id, raid_name, COMPLETED,
	)

	return err
}

// I need functions that will be used to query certain data FROM the database.
