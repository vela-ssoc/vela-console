# console
跟vela security framework 交互的内部接口 需要认证

## 配置
```lua
vela.console{
    network = "unix",
    address = "/tmp/vela.sock",
    script = "resource/script"  --脚本寻找目录
}
```

## 常用方法和命令
console的配置主要有三种模式 根模式(~>>) , CODE模式(chunk>>) , 对象模式(chunk.object>>)
三种模式下的问题 可以使用的命令有区别
### auth
- 认证命令  认证客户端是否是正常用户 只能工作在根模式
- 只能工作在根模式
```bash
vela-go# ./vela-cli unix:///tmp/vela.sock 
~>> auth ****
```

### help 或者 ?
- 用来查看帮助信息
- 所有模式支持
```bash
vela-go# ./vela-cli
~>> help
......
~>> ?
```

### show 或者 list 
- 所有模式都能支持 
- 展示当前的配置文件
```bash
vela-go# ./vela-cli
~>> auth ****
~>> show
......
~>> list
```

### load
- 加载配置代码
- 只工作在 根模式
```bash
vela-go# ./vela-cli
~>> load fasthttp
........
~>> load kafka
```
### use
- 切换工作模式
- 根模式 和 CODE模式
```bash
vela-go# ./vela-cli
~>> use fasthttp
fasthttp>> use kafka
fasthttp.kafka>> show
.....
fasthttp.kafka>> quit
fasthttp>>
```

### .key
- 主要是调用对象内部操作 , 一般是函数调用或者赋值
- 只工作对象模式
```bash
fasthttp.kafka>>.timeout = 100
fasthttp.kafka>>.push("hello")
.......
```

## 如何开发一个满足console的对象
- 1. 首先只用采用L.NewVelaData(string)方法生成的对象(lightUserData)
- 2. 实现 show(out lua.Console) 和 Help(out lua.Console) 方法 ,调用show 和help命令的入口
- 3. 其中.key 或者 .func 取决与 对象中的.Index方法
- 4. 如果想在console下输出 可以才用 CheckPrinter(L) 来获取 console的输出对象
```go
type config struct {
    Name  string       `lua:"name"    type:"string"`
    Age   int          `lua:"age,18"  type:"int"`
    Sdk   lua.Writer   `lua:"sdk"     type:"object"`
}

func (cfg *config) Push(L *lua.LState) int {
    val := L.CheckString(1)
    out := lua.CheckPrinter(L) //注意此时触发事件的入口必须是console
    out.Printf("%s" , val)
    return 0
}

func (cfg *config)Index(L *lua.LState , key string) lua.LValue {
    if key == "push" { return L.NewFunction(cfg.Push )}  //对象模式可以.push("hello")
}

func newConfig(L *lua.LState) *config {
    tab := L.CheckTable(1)
    cfg := &config{}
    //todo
    return cfg
}

func newLuaDemo(L *lua.LState) int {
    cfg := newConfig(L)
    proc := L.NewVelaData(cfg.name) //生成对象
    L.Push(proc)
    return 1
}

func Constructor(env xcall.Env) {
    env.Set("demo" , lua.NewFunction(newLuaDemo))
}
```