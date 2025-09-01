package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func InitLogger() (*zap.Logger, error) {

	log, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("cannot init logger err: %w", err)
	}

	return log, nil

}
