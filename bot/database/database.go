package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"

	"github.com/bwmarrin/discordgo"
)

// Globally accessible
var Db *sql.DB

const (
	StatusCreator    int64 = 0
	StatusRaidLeader int64 = 1
	StatusNoClears   int64 = 2
	StatusReserved   int64 = 3
	StatusActive     int64 = 4
	StatusStandby    int64 = 5
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
func RosterCreateDatabaseEntry(raid_name string, author_id string, note string) (int64, error) {
	// The ID auto-increments and does not need to be included.
	const ACTIVE int = 1
	resp, err := Db.Exec(`
		INSERT INTO ROSTER (raid_name, author_id, is_active, note)
		VALUES (?, ?, ?, ?)`,
		raid_name, author_id, ACTIVE, note)

	if err != nil {
		return 0, fmt.Errorf("Failed to insert new value into database: %s", err)
	}

	roster_id, err := resp.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Failed to retrieve id of last inserted element: %s", err)
	}

	const POSITION_FIRST int = 0
	// This should call the "add to user" function
	resp, err = Db.Exec(`
		INSERT INTO ROSTER_PARTICIPANTS (roster_id, user_id, status, position)
		VALUES (?, ?, ?, ?)`,
		roster_id, author_id, StatusCreator, POSITION_FIRST)

	if err != nil {
		return 0, fmt.Errorf("Failed to add user to roster: %s", err)
	}

	return roster_id, nil
}

func RosterUpdateMessageId(message_id string, roster_id int64) error {
	// log.Printf("Roster ID: %d\n", roster_id)

	_, err := Db.Exec(`UPDATE ROSTER SET message_id = (?) WHERE id = (?)`,
		message_id, roster_id)

	// NOTE: There should probably be an error check

	// return err

	// rows, err := resp.RowsAffected()
	// _, err = resp.RowsAffected()
	// if err != nil {
	// 	return err
	// }

	return err

	// if rows == 0 {
	// 	log.Println("No lines updated.")
	// } else {
	// 	log.Println("Message id successfully updated.")
	// }

	// return nil
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

// NOTE: The func name will have to be renamed - it adds and updates.
// func RosterUserUpsert(roster_id int64, user_id string, position float64) error {
// NOTE: rosterEditLeader() and rosterEditReserve() would prefer to use raid name
// over roster_id.
func RosterUserUpsert(roster_id int64, user_id string, status int64) error {
	// Potential errors:
	// - user already exists in roster
	// - roster is archived (handled by the message being deleted when inactive)

	// NOTE: This must dynamically pull if the user has completed the raid and
	// auto-apply the proper status. Then, it should grab the largest number of
	// that status and set the position value as +1.

	// Get the raid name, based on the roster id
	raidName, err := RaidQueryNameFromRosterId(roster_id)

	// Get user completion status, based on raid name
	if status == StatusActive {
		hasCompleted, err := QueryUserRaidCompletion(user_id, raidName)
		if hasCompleted == false {
			status = StatusNoClears
		}
		if err != nil {
			log.Println(err)
		}
	}

	// Get the highest position of those in this status
	position, err := RosterParticipantsQueryNextPosition(roster_id, status)
	// NOTE: We need an error check, for broken requests.

	// NOTE: UPSERT
	_, err = Db.Exec(`
		INSERT INTO ROSTER_PARTICIPANTS (roster_id, user_id, status, position)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (roster_id, user_id)
		DO UPDATE SET status = excluded.status`,
		roster_id, user_id, status, position,
	)

	if err != nil {
		return fmt.Errorf("Failed to add user to roster: %s", err)
	}

	return nil
}

func RosterUserLeave(roster_id int64, user_id string) error {
	// Potential errors:
	// - user is not in roster
	// - roster is archived
	// - author is trying to leave with no other players in roster
	_, err := Db.Exec(`
		DELETE FROM ROSTER_PARTICIPANTS
		WHERE roster_id = ? AND user_id = ?`,
		roster_id, user_id,
	)

	if err != nil {
		return fmt.Errorf("Failed to remove user from roster: %s", err)
	}

	return nil
}

// Obsolete
func RosterUpdateRaidLeader(roster_id int64, previous_leader_id string,
	new_leader_id string) error {

	// Check if current raid leader == user making request.

	// Check if new raid leader exists in this roster.

	return nil
}

// Obsolete
func RosterReserveUser(roster_id int64, user_id string) error {
	// Potential errors:
	// - user does not exist in roster
	// - roster is archived
	// - caller is not roster author
	return nil
}

type Roster struct {
	ID       int64
	RaidName string
	AuthorID string
	// Must be pointer, since can be returned before value is set
	MessageID *string
	Note      *string
}

func RosterQueryRowFromRaidName(raid_name string) (*Roster, error) {
	var roster Roster
	const ACTIVE int = 1

	err := Db.QueryRow(`
		SELECT id, raid_name, author_id, message_id, note
		FROM ROSTER
		WHERE raid_name = ? AND is_active = ?`,
		raid_name, ACTIVE,
	).Scan(&roster.ID, &roster.RaidName, &roster.AuthorID, &roster.MessageID, &roster.Note)

	if err != nil {
		return nil, fmt.Errorf("No active roster found (%s).", err)
	}

	return &roster, nil
}

func QueryUserRaidCompletion(userID string, raidName string) (bool, error) {
	var hasCompleted bool

	row := Db.QueryRow(`
		SELECT has_completed
		FROM USER_RAID_COMPLETIONS
		WHERE user_id = ? AND raid_name = ?`,
		userID, raidName,
	)

	err := row.Scan(&hasCompleted)
	if err != nil {
		return false, err
	}

	return hasCompleted, nil
}

type Raid struct {
	Name        string
	Description string
	RoleID      string
	InRotation  bool
}

// NOTE: This is currently breaking the button functionality
func RaidQueryRowFromName(raid_name string) (*Raid, error) {
	var raid Raid

	err := Db.QueryRow(`
		SELECT name, description, dlc_role_id, in_rotation
		FROM RAID
		WHERE name = ?`,
		raid_name,
	).Scan(&raid.Name, &raid.Description, &raid.RoleID, &raid.InRotation)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve raid information (%s).", err)
	}

	return &raid, nil
}

func RaidQueryNameFromRosterId(rosterID int64) (string, error) {
	var raidName string

	row := Db.QueryRow(`
		SELECT raid_name
		FROM ROSTER
		WHERE id = ?`,
		rosterID,
	)

	err := row.Scan(&raidName)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return raidName, nil
}

// func RosterQueryIdFromUser(raid_name string, author_id string) (int64, error) {
// 	var rosterId int64
// 	const ACTIVE int = 1
//
// 	err := Db.QueryRow(`
// 		SELECT id
// 		FROM ROSTER
// 		WHERE raid_name = ? AND author_id = ? AND is_active = ?`,
// 		raid_name, author_id, ACTIVE,
// 	).Scan(&rosterId)
//
// 	if err != nil {
// 		return 0, err
// 	}
//
// 	return rosterId, nil
// }

func RosterParticipantsQueryNextPosition(rosterID int64, status int64) (float64, error) {
	var position sql.NullFloat64

	row := Db.QueryRow(`
		SELECT MAX(position)
		FROM ROSTER_PARTICIPANTS
		WHERE status = ?`,
		status,
	)

	err := row.Scan(&position)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	if !position.Valid {
		return 0, nil
	}

	return position.Float64, nil
}

func RosterQueryIdFromMessageId(message_id string) (int64, error) {
	const IS_ACTIVE int = 1

	var rosterID int64
	row := Db.QueryRow(`
		SELECT id
		FROM ROSTER
		WHERE message_id = ? AND is_active = ?`,
		message_id, IS_ACTIVE,
	)

	err := row.Scan(&rosterID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No roster was found (%s)\n", err)
			return 0, err
		}
	}

	return rosterID, nil
}

// NOTE: Finish implementing this.
func RosterQueryRaidNameFromId(roster_id int64) (string, error) {
	var raidName string

	row := Db.QueryRow(`
		SELECT raid_name
		FROM ROSTER
		WHERE id = ?`,
		roster_id,
	)

	err := row.Scan(&raidName)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return raidName, nil
}

// NOTE: Modify to only grab users of (up to) 'status active' and first <=6.
// If other functions use this, instead make a new function for this.
func RosterQueryUserIds(roster_id int64) []string {
	var ids []string

	rows, err := Db.Query(`
		SELECT user_id
		FROM ROSTER_PARTICIPANTS
		WHERE roster_id = ?`,
		roster_id,
	)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string

		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return ids
}

type User struct {
	ID     string
	Status int64
}

type RosterGroup struct {
	ActiveGroup  []User
	StandbyGroup []User
}

// func RosterQueryUserIdsSorted(roster_id int64) []User {
func RosterQueryUserIdsSorted(roster_id int64) RosterGroup {
	var activeGroup []User
	var standbyGroup []User
	// var rosterGroup RosterGroup

	// var users []User

	const MAX_RAID_SIZE = 6

	rows, err := Db.Query(`
		SELECT user_id, status
		FROM ROSTER_PARTICIPANTS
		WHERE roster_id = ?
		ORDER BY status ASC, position ASC`,
		roster_id,
	)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var status int64

		if err := rows.Scan(&id, &status); err != nil {
			log.Fatal(err)
		}
		// users = append(users, User{id, status})
		if len(activeGroup) >= MAX_RAID_SIZE || status == StatusStandby {
			// rosterGroup.StandbyGroup = append(rosterGroup.StandbyGroup, User{id, status})
			// group2 = append(group2, User{id, status})
			standbyGroup = append(standbyGroup, User{id, status})
		} else {
			// rosterGroup.ActiveGroup = append(rosterGroup.ActiveGroup, User{id, status})
			// group1 = append(group2, User{id, status})
			activeGroup = append(activeGroup, User{id, status})
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// return users
	// return RosterGroup{group1, group2}
	// return rosterGroup
	return RosterGroup{activeGroup, standbyGroup}
}

func MarkUserRaidCompletion(user_id string, raid_name string) error {
	const COMPLETED int = 1
	_, err := Db.Exec(`
		INSERT INTO USER_RAID_COMPLETIONS (user_id, raid_name, has_completed)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id, raid_name)
		DO UPDATE SET has_completed = excluded.has_completed`,
		user_id, raid_name, COMPLETED,
	)

	return err
}

// I need functions that will be used to query certain data FROM the database.
