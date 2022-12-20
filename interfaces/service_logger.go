package interfaces

//ServiceLogger is interface for loggers
type ServiceLogger interface {
	Info(string)
	Warning(string)
	Error(string)
}
