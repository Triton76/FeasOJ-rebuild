package global

// 判题状态常量
const (
	// 正在运行
	Running string = "judging"
	// 答案正确
	Accepted string = "accepted"
	// 答案错误
	WrongAnswer string = "wrong_answer"
	// 编译错误
	CompileError string = "compile_error"
	// 超出时间限制
	TimeLimitExceeded string = "time_limit_exceeded"
	// 超出内存限制
	MemoryLimitExceeded string = "memory_limit_exceeded"
	// 运行时错误
	RuntimeError string = "runtime_error"
	// 系统错误
	SystemError string = "system_error"
)
