package sync_test

import (
	"testing"

	"github.com/hephbuild/heph/internal/hlsp/capabilities/sync"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ParseNewBytesSuite struct {
	suite.Suite
}

func (s *ParseNewBytesSuite) TestInsert() {
	currBytes := []byte("def hello_world():")
	newBytes := []byte("_my_precious")
	result := sync.ParseNewBytes(currBytes, newBytes, 15, 15)
	s.Require().Equal( "def hello_world_my_precious():", string(result))
}

func (s *ParseNewBytesSuite) TestInsertWithNewLine() {
	currBytes := []byte("def hello_world():")
	newBytes := []byte("_my_precious\n")
	result := sync.ParseNewBytes(currBytes, newBytes, 15, 15)
	s.Require().Equal( "def hello_world_my_precious\n():", string(result))
}

func (s *ParseNewBytesSuite) TestReplace() {
	currBytes := []byte("def hello_world():")
	newBytes := []byte("world_hello")
	result := sync.ParseNewBytes(currBytes, newBytes, 4, 4+len(newBytes))
	s.Require().Equal( "def world_hello():", string(result))
}

func (s *ParseNewBytesSuite) TestReplaceWithMore() {
	currBytes := []byte("def hello_friends():")
	newBytes := []byte("hi_world")
	result := sync.ParseNewBytes(currBytes, newBytes, 4, 9)
	s.Require().Equal( "def hi_world_friends():", string(result))
}

func (s *ParseNewBytesSuite) TestReplaceWithFewer() {
	currBytes := []byte("def hello_world():")
	newBytes := []byte("hi")
	result := sync.ParseNewBytes(currBytes, newBytes, 4, 9)
	s.Require().Equal( "def hi_world():", string(result))
}

func (s *ParseNewBytesSuite) TestRemove() {
	currBytes := []byte("def hello_world():")
	newBytes := []byte("")
	result := sync.ParseNewBytes(currBytes, newBytes, 4, 10)
	s.Require().Equal( "def world():", string(result))
}

func TestParseNewBytes(t *testing.T) {
	suite.Run(t, new(ParseNewBytesSuite))
}
