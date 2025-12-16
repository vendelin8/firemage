package testutil

import (
	"errors"

	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/log"
	"go.uber.org/zap"
)

var ErrMock = errors.New("mock error")

func InitLog() func() error {
	logger, _ := zap.NewDevelopment()
	log.Lgr = logger
	return logger.Sync
}

func BuildActionsMap(data map[string]map[string]any) map[string]common.ClaimsMap {
	result := make(map[string]common.ClaimsMap)
	for uid, actionMap := range data {
		cm, _ := common.NewClaimsMapFrom(actionMap)
		if cm != nil {
			result[uid] = *cm
		}
	}
	return result
}
