package contexts

type ContextKey string

const (
	ContextKey__TraceId      ContextKey = "TraceId"
	ContextKey__TimeMs       ContextKey = "TimeMs"
	ContextKey__ClientId     ContextKey = "ClientId"
	ContextKey__AppSubjectId ContextKey = "AppSubjectId"
	ContextKey__MsSubjectId  ContextKey = "MsSubjectId"
)
