# time

基础时间处理库

- [time.sec()]()
- [time.min()]()
- [time.hour()]()
- [time.day()]()
- [time.week()]()
- [time.month()]()
- [time.year()]()
- [time.sleep(v)]()
- [time.now(v)]()
- [time.at(v)]()
- [time.today()]()

## example
```lua
    local sec = time.sec()
    print(sec) -- 59

    local min = time.min()
    print(min) -- 58

    local hour = time.hour()
    print(hour) -- 01 或 13

    local day = time.day()
    print(day)  -- 30

    local week = time.week()
    print(week) -- Sunday

    local month = time.month()
    print(month) -- 12

    local year = time.year()
    print(year)  -- 2022

    time.sleep(1 * 1000) -- sleep 1s

    local td = time.today() -- 今天
    print(td) -- 2022-02-22

    local td = time.today(1) -- 一天
    print(td) -- 2022-02-23

    local td = time.today(-1) -- 前一天时间
    print(td) -- 2022-02-21
```
