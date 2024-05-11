package database

import (
	"math"
	"testing"

	"github.com/capdale/was/config"
	"github.com/capdale/was/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}

type DatabaseSuite struct {
	suite.Suite
	d      *DB
	tmpDir test.TmpDir
}

func (s *DatabaseSuite) BeforeTest(suiteName string, testname string) {
	s.tmpDir = test.NewTmpDir("was_database")
	db, err := NewSQLite(&config.SQLite{
		Path: s.tmpDir.Join("test.db"),
	}, math.MaxInt)
	assert.Nil(s.T(), err)
	s.d = db
}

func (s *DatabaseSuite) AfterTest(suiteName string, testname string) {
	err := s.d.Close()
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), s.tmpDir.Close())
}
