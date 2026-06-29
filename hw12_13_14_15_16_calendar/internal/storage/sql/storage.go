package sqlstorage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
)

const uniqueViolationCode = "23505"

var (
	ErrNotConnected = errors.New("database is not connected")
)

type SqlStorage struct {
	dataSourceName string
	db             *sql.DB
}

func New(dataSourceName string) *SqlStorage {
	if dataSourceName == "" {
		dataSourceName = os.Getenv("DATABASE_URL")
	}

	return &SqlStorage{dataSourceName: dataSourceName}
}

func (s *SqlStorage) Connect() error {
	if s.db != nil {
		return s.db.Ping()
	}
	if s.dataSourceName == "" {
		return fmt.Errorf("%w: empty dsn", ErrNotConnected)
	}

	db, err := sql.Open("pgx", s.dataSourceName)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}

	s.db = db
	return nil
}

func (s *SqlStorage) Close() error {
	if s.db == nil {
		return nil
	}

	err := s.db.Close()
	s.db = nil
	return err
}

func (s *SqlStorage) CreateEvent(event storage.Event) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	tx, err := s.beginTx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

	if err := s.ensureDateAvailable(tx, event, uuid.Nil); err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO events (
			id,
			title,
			date,
			duration,
			description,
			user_id,
			notify_before
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		event.ID,
		event.Title,
		event.Date,
		event.Duration.Nanoseconds(),
		nullableString(event.Description),
		event.UserID,
		nullableDuration(event.NotifyBefore),
	)
	if isUniqueViolation(err) {
		return storage.ErrEventAlreadyExists
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SqlStorage) UpdateEvent(id uuid.UUID, event storage.Event) error {
	tx, err := s.beginTx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

	if err := ensureEventExists(tx, id); err != nil {
		return err
	}

	event.ID = id
	if err := s.ensureDateAvailable(tx, event, id); err != nil {
		return err
	}

	result, err := tx.Exec(`
		UPDATE events
		SET
			title = $2,
			date = $3,
			duration = $4,
			description = $5,
			user_id = $6,
			notify_before = $7
		WHERE id = $1
	`,
		id,
		event.Title,
		event.Date,
		event.Duration.Nanoseconds(),
		nullableString(event.Description),
		event.UserID,
		nullableDuration(event.NotifyBefore),
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return storage.ErrEventNotFound
	}

	return tx.Commit()
}

func (s *SqlStorage) DeleteEvent(id uuid.UUID) error {
	db, err := s.connection()
	if err != nil {
		return err
	}

	result, err := db.Exec(`DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *SqlStorage) ListEventsForDay(date time.Time) ([]storage.Event, error) {
	start := dayStart(date)
	return s.listEventsBetween(start, start.AddDate(0, 0, 1))
}

func (s *SqlStorage) ListEventsForWeek(startOfWeek time.Time) ([]storage.Event, error) {
	start := dayStart(startOfWeek)
	return s.listEventsBetween(start, start.AddDate(0, 0, 7))
}

func (s *SqlStorage) ListEventsForMonth(startOfMonth time.Time) ([]storage.Event, error) {
	start := dayStart(startOfMonth)
	return s.listEventsBetween(start, start.AddDate(0, 1, 0))
}

func (s *SqlStorage) listEventsBetween(start, end time.Time) ([]storage.Event, error) {
	db, err := s.connection()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT
			id,
			title,
			date,
			duration,
			description,
			user_id,
			notify_before
		FROM events
		WHERE date >= $1 AND date < $2
		ORDER BY date, id
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]storage.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (s *SqlStorage) ensureDateAvailable(tx *sql.Tx, event storage.Event, excludedID uuid.UUID) error {
	var busy bool
	err := tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM events
			WHERE
				user_id = $1
				AND id <> $2
				AND date < $3
				AND date + make_interval(secs => duration::double precision / 1000000000) > $4
		)
	`,
		event.UserID,
		excludedID,
		event.Date.Add(event.Duration),
		event.Date,
	).Scan(&busy)
	if err != nil {
		return err
	}
	if busy {
		return storage.ErrDateBusy
	}

	return nil
}

func (s *SqlStorage) beginTx() (*sql.Tx, error) {
	db, err := s.connection()
	if err != nil {
		return nil, err
	}

	return db.Begin()
}

func (s *SqlStorage) connection() (*sql.DB, error) {
	if s.db == nil {
		return nil, ErrNotConnected
	}

	return s.db, nil
}

func ensureEventExists(tx *sql.Tx, id uuid.UUID) error {
	var exists bool
	err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM events WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return storage.ErrEventNotFound
	}

	return nil
}

func scanEvent(scanner interface {
	Scan(dest ...any) error
}) (storage.Event, error) {
	var (
		idRaw           string
		userIDRaw       string
		durationRaw     int64
		descriptionRaw  sql.NullString
		notifyBeforeRaw sql.NullInt64
		event           storage.Event
	)

	if err := scanner.Scan(
		&idRaw,
		&event.Title,
		&event.Date,
		&durationRaw,
		&descriptionRaw,
		&userIDRaw,
		&notifyBeforeRaw,
	); err != nil {
		return storage.Event{}, err
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return storage.Event{}, err
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return storage.Event{}, err
	}

	event.ID = id
	event.UserID = userID
	event.Duration = time.Duration(durationRaw)
	if descriptionRaw.Valid {
		event.Description = &descriptionRaw.String
	}
	if notifyBeforeRaw.Valid {
		notifyBefore := time.Duration(notifyBeforeRaw.Int64)
		event.NotifyBefore = &notifyBefore
	}

	return event, nil
}

func nullableString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}

	return sql.NullString{String: *value, Valid: true}
}

func nullableDuration(value *time.Duration) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}

	return sql.NullInt64{Int64: value.Nanoseconds(), Valid: true}
}

func dayStart(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func rollbackTx(tx *sql.Tx) {
	_ = tx.Rollback()
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}
