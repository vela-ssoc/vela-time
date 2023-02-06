package vtime

import (
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"github.com/vela-ssoc/vela-kit/worker"
	"gopkg.in/tomb.v2"
	"time"
)

type VTime time.Time

func (vt VTime) Type() lua.LValueType                   { return lua.LTObject }
func (vt VTime) AssertFloat64() (float64, bool)         { return vt.toF(), true }
func (vt VTime) AssertString() (string, bool)           { return "", false }
func (vt VTime) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (vt VTime) Peek() lua.LValue                       { return vt }

func (vt VTime) String() string {
	tt := time.Time(vt)
	return tt.Format(time.RFC3339Nano)
}

func (vt VTime) toF() float64 {
	tt := time.Time(vt)
	return float64(tt.UnixNano())
}

func (vt VTime) format(L *lua.LState) int {
	tt := time.Time(vt)
	fm := L.CheckString(1)
	L.Push(lua.S2L(tt.Format(fm)))
	return 1
}

func (vt VTime) Index(L *lua.LState, key string) lua.LValue {
	tt := time.Time(vt)

	switch key {
	case "sec":
		return lua.LInt(tt.Second())
	case "min":
		return lua.LInt(tt.Minute())
	case "hour":
		return lua.LInt(tt.Hour())
	case "day":
		return lua.LInt(tt.Day())
	case "week":
		return lua.S2L(tt.Weekday().String())
	case "month":
		return lua.LInt(tt.Month())
	case "year":
		return lua.LInt(tt.Year())
	case "format":
		return L.NewFunction(vt.format)
	case "tt_sec":
		return lua.LNumber(tt.Unix())
	case "tt_milli":
		return lua.LNumber(tt.UnixMilli())
	case "tt_nano":
		return lua.LNumber(tt.UnixNano())
	case "today":
		return lua.S2L(tt.Format("2006-01-02"))
	default:
		return lua.LNil
	}
}

func newLuaDay(L *lua.LState) int {
	L.Push(lua.LInt(time.Now().Day()))
	return 1
}
func newLuaYear(L *lua.LState) int {
	L.Push(lua.LInt(time.Now().Year()))
	return 1
}
func newLuaHour(L *lua.LState) int {
	L.Push(lua.LInt(time.Now().Year()))
	return 1
}

func newLuaWeek(L *lua.LState) int {
	L.Push(lua.S2L(time.Now().Weekday().String()))
	return 1
}
func newLuaMonth(L *lua.LState) int {
	L.Push(lua.LInt(time.Now().Month()))
	return 1
}
func newLuaMinute(L *lua.LState) int {
	L.Push(lua.LInt(time.Now().Minute()))
	return 1
}

func newLuaSecond(L *lua.LState) int {
	L.Push(lua.LInt(time.Now().Second()))
	return 1
}

func newLuaTimeSleep(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	lv := L.Get(1)
	delay, ok := lv.AssertFloat64()
	if !ok {
		return 0
	}

	<-time.After(time.Duration(delay) * time.Millisecond)

	return 0
}

func newLuaTimeNow(L *lua.LState) int {
	val := L.Get(1)
	now := time.Now()

	switch val.Type() {

	case lua.LTString:
		fm := val.String()
		switch fm {
		case "mil":
			L.Push(lua.S2L(now.Format("2006-01-02 15:04:05.00")))
			return 1
		case "sec":
			L.Push(lua.S2L(now.Format("2006-01-02 15:04:05")))
			return 1
		case "min":
			L.Push(lua.S2L(now.Format("2006-01-02 15:04")))
			return 1
		case "hour":
			L.Push(lua.S2L(now.Format("2006-01-02.15")))
			return 1
		case "day":
			L.Push(lua.S2L(now.Format("2006-01-02")))
			return 1

		case "mon":
			L.Push(lua.S2L(now.Format("2006-01")))
			return 1
		case "year":
			L.Push(lua.S2L(now.Format("2006")))
			return 1

		default:
			L.Push(lua.S2L(now.Format(fm)))
			return 1
		}

	default:
		L.Push(VTime(time.Now()))
		return 1
	}
}

func newluaTimeAt(L *lua.LState) int {
	delay := L.IsInt(1)
	pip := pipe.NewByLua(L, pipe.Seek(1))
	if pip.Len() == 0 {
		return 0
	}

	now := time.Now().Unix()
	xEnv.Spawn(0, func() {
		rt := <-time.After(time.Duration(delay) * time.Millisecond)
		pip.Do(time.Now().Unix(), xEnv.Clone(L), func(err error) {
			audit.Errorf("延时执行失败 start:%d delay:%d run:%d error %v", now, delay, rt.Unix(), err).
				Log().From(L.CodeVM()).Put()
		})
		audit.Errorf("延时执行结束 start:%d delay:%d run:%d", now, delay, rt.Unix()).
			Log().From(L.CodeVM()).Put()
	})
	return 0
}

func newLuaTimeEvery(L *lua.LState) int {
	if L.CodeVM() == "" {
		L.RaiseError("time every must task code run")
		return 0
	}
	pip := pipe.NewByLua(L, pipe.Seek(1))
	if pip.Len() == 0 {
		return 0
	}

	interval := L.IsInt(1)
	if interval <= 0 {
		interval = 10
	}

	tk := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer tk.Stop()

	tom := new(tomb.Tomb)
	task := func() {
		for {
			select {
			case <-tom.Dying():
				return
			case now := <-tk.C:
				pip.Do(now, xEnv.Clone(L), func(err error) {
					audit.Errorf("周期执行失败 run:%d error %v", err).
						Log().From(L.CodeVM()).Put()
				})
			}

		}
	}

	kill := func() {
		tom.Kill(nil)
	}

	wk := worker.New(L, "time.every."+auxlib.ToString(time.Now().Unix())).Task(task).Kill(kill)

	xEnv.Start(L, wk).From(L.CodeVM()).Do()
	return 0
}

func newLuaTimeToday(L *lua.LState) int {
	n := L.IsInt(1)
	if n == 0 {
		L.Push(lua.S2L(time.Now().Format("2006-01-02")))
		return 1
	}
	L.Push(lua.S2L(time.Now().AddDate(0, 0, n).Format("2006-01-02")))
	return 1
}

func newLuaTimeParse(L *lua.LState) int {
	layout := L.CheckString(1)
	val := L.CheckString(2)

	tv, err := time.Parse(layout, val)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.S2L(err.Error()))
		return 2
	}

	L.Push(VTime(tv))
	return 1
}

func New(t time.Time) VTime {
	return VTime(t)
}
