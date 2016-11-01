package failure

// 失败
type Failure struct {
	// 代码
	Code   string

	// 详情
	Detail interface{}
}

// 创建一个失败
func New(code string) *Failure {
	return NewWithDetail(code, nil)
}

// 创建一个失败
func NewWithDetail(code string, detail interface{}) *Failure {
	return &Failure{
		Code: code,
		Detail: detail,
	}
}

func (fail *Failure) Error() string {
	return fail.Code
}
