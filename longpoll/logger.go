package longpoll

import "go.uber.org/zap"

func NewLogger() *zap.SugaredLogger {
	cfg := zap.Config{
		Level:    zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding: "json",
		OutputPaths: []string{
			"stdout",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}

	logger, _ := cfg.Build()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	return sugar
}
