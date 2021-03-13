package state

import "github.com/eagraf/habitat-node/entities"

type LogCollection struct {
	CommunityLogs map[entities.CommunityID]*Log
	NodeLog       *Log
}

type Log struct {
}

type Entry struct {
	TransitionCategory *entities.TransitionCategory
	Transition         *entities.Transition
}
