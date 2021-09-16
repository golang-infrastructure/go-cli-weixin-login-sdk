package logger

func Info(format string, s ...interface{}) {
	White(format + "\n", s...)
}

func Error(format string, s ...interface{}) {
	Red(format + "\n", s...)
}

func Success(format string, s ...interface{}) {
	Green(format + "\n", s...)
}
