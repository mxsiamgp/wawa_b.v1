package rest_json_rpc

import (
	"net/http"
	"sync"

	"wawa_b.v1/module/rest_json_rpc/failure"

	"github.com/labstack/echo"
)

// 状态代码
const (
	// 成功
	STAT_CD_OK = "OK"

	// 失败
	STAT_CD_FAIL = "FAIL"
)

// 失败代码
const (
	// 过程不存在
	FAIL_CD_NO_SUCH_PROCESS = "REST_JSON_RPC.NO_SUCH_PROCESS"

	// 没有指定过程名
	FAIL_CD_NOT_SPECIFIED_PROCESS = "REST_JSON_RPC.NOT_SPECIFIED_PROCESS"
)

type ProcessChain struct {
	// 过程
	*Process

	// Echo上下文
	context             echo.Context

	// 当前处理器索引
	currentHandlerIndex int

	// 参数
	param               interface{}
}

// 过程处理器
type ProcessHandler func(ctx echo.Context, param interface{}, ch *ProcessChain) interface{}

// 过程参数工厂
type ProcessParamFactory func() interface{}

// 过程
type Process struct {
	//处理器
	Handlers     []ProcessHandler

	//参数工厂
	ParamFactory ProcessParamFactory
}

// 远程过程调用
type RPC struct {
	//用于保证processes的并发安全
	mutex     *sync.RWMutex

	//已注册的过程
	processes map[string]*Process
}

func (ch *ProcessChain) Next() interface{} {
	if ch.currentHandlerIndex == len(ch.Process.Handlers) - 1 {
		return nil
	}
	ch.currentHandlerIndex++
	return ch.Process.Handlers[ch.currentHandlerIndex](ch.context, ch.param, ch)
}

// 创建一个远程过程调用
func NewRPC() *RPC {
	return &RPC{
		mutex: &sync.RWMutex{},
		processes: make(map[string]*Process),
	}
}

// 注册一个过程
// 过程命名规范：[模块名].[过程名]
func (rpc *RPC) RegisterProcess(name string, proc *Process) {
	rpc.mutex.Lock()
	defer rpc.mutex.Unlock()
	rpc.processes[name] = proc
}

// 反注册一个过程
func (rpc *RPC) UnregisterProcess(name string) {
	rpc.mutex.Lock()
	defer rpc.mutex.Unlock()
	delete(rpc.processes, name)
}

// 根据名称获取过程
func (rpc *RPC) GetProcess(name string) (*Process, bool) {
	rpc.mutex.RLock()
	defer rpc.mutex.RUnlock()
	proc, ok := rpc.processes[name]
	return proc, ok
}

// 获取Echo处理器
func (rpc *RPC) HandlerFunc() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		name := ctx.QueryParam("process")
		if len(name) == 0 {
			fail(ctx, FAIL_CD_NOT_SPECIFIED_PROCESS, nil)
			return nil
		}

		proc, ok := rpc.GetProcess(name)
		if !ok {
			fail(ctx, FAIL_CD_NO_SUCH_PROCESS, nil)
			return nil
		}

		param := proc.ParamFactory()
		if err := ctx.Bind(param); err != nil {
			return err
		}

		//失败处理
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(*failure.Failure)
				if !ok {
					panic(r)
				}

				fail(ctx, err.Code, err.Detail)
			}
		}()

		result := proc.Handlers[0](ctx, param, &ProcessChain{
			Process: proc,
			context: ctx,
			currentHandlerIndex: 0,
			param: param,
		})

		ctx.JSON(http.StatusOK, map[string]interface{}{
			"status_code": STAT_CD_OK,
			"result": result,
		})

		return nil
	}
}

func fail(ctx echo.Context, code string, detail interface{}) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status_code": STAT_CD_FAIL,
		"fail_code": code,
		"fail_detail": detail,
	})
}
