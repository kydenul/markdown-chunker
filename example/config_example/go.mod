module config_example

go 1.24.5

replace github.com/kydenul/markdown-chunker => ../../

require github.com/kydenul/markdown-chunker v0.0.0-00010101000000-000000000000

require (
	github.com/kydenul/log v1.2.0 // indirect
	github.com/yuin/goldmark v1.7.13 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
