// Package drivers (database/drivers) contains database
// driver structs for accessing various database types
//   Authors: Ringo Hoffmann
package drivers

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gorilla/sessions"
	"github.com/zekroTJA/mysqlstore"
	"github.com/zekroTJA/vplan2019/internal/database"
	"github.com/zekroTJA/vplan2019/pkg/multierror"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

// MySQL contains database functions
// for MySQL database
type MySQL struct {
	cfg   map[string]string
	dsn   string
	db    *sql.DB
	stmts *prepStatements
}

type prepStatements struct {
	selectAPITokenByToken *sql.Stmt
	selectAPITokenByIdent *sql.Stmt
	updateAPIToken        *sql.Stmt
	insertAPIToken        *sql.Stmt
	deleteAPIToken        *sql.Stmt

	selectVPlans              *sql.Stmt
	selectVPlanEntries        *sql.Stmt
	selectVPlanEntriesByClass *sql.Stmt
}

// Connect opens a MySql3 database file or creates
// it if it does not exist depending on the passed options
func (s *MySQL) Connect(options map[string]string) error {
	var err error

	s.cfg = options
	s.dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s",
		options["user"], options["password"], options["host"], options["database"])

	s.db, err = sql.Open("mysql", s.dsn)
	err = s.setupPrepStatements()

	return err
}

func (s *MySQL) prepareStatement(multiError *multierror.MultiError, query string) *sql.Stmt {
	stmt, err := s.db.Prepare(query)
	multiError.Append(err)
	return stmt
}

func (s *MySQL) setupPrepStatements() error {
	s.stmts = new(prepStatements)
	m := multierror.NewMultiError(nil)

	s.stmts.selectAPITokenByToken = s.prepareStatement(m, "SELECT ident, expire FROM apitoken WHERE token = ?")
	s.stmts.selectAPITokenByIdent = s.prepareStatement(m, "SELECT token, expire FROM apitoken WHERE ident = ?")
	s.stmts.updateAPIToken = s.prepareStatement(m, "UPDATE apitoken SET token = ?, expire = ? WHERE ident = ?")
	s.stmts.insertAPIToken = s.prepareStatement(m, "INSERT INTO apitoken (ident, token, expire) VALUES (?, ?, ?)")
	s.stmts.deleteAPIToken = s.prepareStatement(m, "DELETE FROM apitoken WHERE ident = ?")

	s.stmts.selectVPlans = s.prepareStatement(m,
		"SELECT id, date_edit, date_for, block, header, footer FROM vplan WHERE "+
			"date_for >= ? AND deleted = 0")
	s.stmts.selectVPlanEntries = s.prepareStatement(m,
		"SELECT id, vplan_id, class, time, messures, responsible FROM vplan_details WHERE vplan_id = ? AND deleted = 0")
	s.stmts.selectVPlanEntriesByClass = s.prepareStatement(m,
		"SELECT id, vplan_id, class, time, messures, responsible FROM vplan_details WHERE vplan_id = ? AND class = ? AND deleted = 0")

	return m.Concat()
}

// Close closes the MySql3 database file
func (s *MySQL) Close() {
	s.db.Close()
}

// Setup creates tables if they do not exist yet
func (s *MySQL) Setup() error {
	_, err := s.db.Exec("CREATE TABLE IF NOT EXISTS `apitoken` (" +
		"`id` int PRIMARY KEY AUTO_INCREMENT," +
		"`ident` text NOT NULL," +
		"`token` text NOT NULL," +
		"`expire` timestamp NOT NULL );")
	if err != nil {
		return err
	}

	return nil
}

// GetAPIToken returns the matching indent and expire time to a found token.
// If the token could not be matched, this returns an empty string without
// and error. Errors are only returned if the database request failes.
func (s *MySQL) GetAPIToken(token string) (string, time.Time, error) {
	var ident string
	var expire database.Timestamp

	row := s.stmts.selectAPITokenByToken.QueryRow(token)
	err := row.Scan(&ident, &expire)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return "", time.Time{}, err
	}

	if ident == "" {
		return ident, time.Time{}, nil
	}

	tExpire, err := time.Parse(timeFormat, string(expire))

	return ident, tExpire, err
}

// GetUserAPIToken gets the users API token with the time, the token expires,
// if existent. Else, the returned string will be empty. If the query failes,
// an error will be returned
func (s *MySQL) GetUserAPIToken(ident string) (string, time.Time, error) {
	var token string
	var expire int64

	row := s.stmts.selectAPITokenByIdent.QueryRow(ident)
	err := row.Scan(&token, &expire)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return "", time.Time{}, err
	}

	tExpire, err := time.Parse(timeFormat, string(expire))

	return token, tExpire, err
}

// SetUserAPIToken sets the API token an the expire time of it for a user
func (s *MySQL) SetUserAPIToken(ident, token string, expire time.Time) error {
	res, err := s.stmts.updateAPIToken.Exec(token, expire, ident)
	if err != nil {
		return err
	}
	if ar, err := res.RowsAffected(); err != nil {
		return err
	} else if ar < 1 {
		_, err = s.stmts.insertAPIToken.Exec(ident, token, expire)
	}
	return err
}

// DeleteUserAPIToken removes a users token from the database
func (s *MySQL) DeleteUserAPIToken(ident string) error {
	_, err := s.stmts.deleteAPIToken.Exec(ident)
	return err
}

// GetConfigModel returns a map with preset config
// keys and values
func (s *MySQL) GetConfigModel() map[string]string {
	return map[string]string{
		"host":     "localhost",
		"user":     "vplan2",
		"password": "",
		"database": "vplan2",
	}
}

// GetSessionStoreDriver returns a new instance of the session
// store driver, which should be used for saving encrypted session data
func (s *MySQL) GetSessionStoreDriver(maxAge int, secrets ...[]byte) (sessions.Store, error) {
	return mysqlstore.NewMySQLStoreFromConnection(s.db, "apisessions", "/", maxAge, secrets...)
}

// GetVPlans collects VPlans wich for-dates are after the passed timestamp.
// Also, a class can be specified for filtering the VPlanEntries.
func (s *MySQL) GetVPlans(class string, timestamp time.Time) ([]*database.VPlan, error) {
	rows, err := s.stmts.selectVPlans.Query(timestamp)
	if err != nil {
		return nil, err
	}

	vplans := make([]*database.VPlan, 0)
	mErr := multierror.NewMultiError(nil)

	var dateEdit, dateFor database.Timestamp
	for rows.Next() {
		vplan := new(database.VPlan)
		err = rows.Scan(&vplan.ID, &dateEdit, &dateFor, &vplan.Block, &vplan.Header, &vplan.Footer)
		mErr.Append(err)
		if err == nil {
			vplan.DateEdit, _ = time.Parse(timeFormat, string(dateEdit))
			vplan.DateFor, _ = time.Parse(timeFormat, string(dateFor))
			vplans = append(vplans, vplan)
		}
	}

	for _, v := range vplans {
		if class != "" {
			rows, err = s.stmts.selectVPlanEntriesByClass.Query(v.ID, class)
		} else {
			rows, err = s.stmts.selectVPlanEntries.Query(v.ID)
		}
		mErr.Append(err)
		if err != nil {
			continue
		}

		for rows.Next() {
			entry := new(database.VPlanEntry)
			err = rows.Scan(&entry.ID, &entry.VPlanID, &entry.Class, &entry.Time, &entry.Messures, &entry.Resposible)
			mErr.Append(err)
			if err == nil {
				v.Entries = append(v.Entries, entry)
			}
		}
	}

	return vplans, mErr.Concat()
}