
1. logger 
 - 是对zap进行了封装，使用全局单例Sugar Logger
 - gin automaxprocs gorm的log都替换成了zap
 - 如果用户不自己初始化的话，会自动初始化为info级别到标准输出的logger
 - SetUpLog 不会进行日志切分轮转
 - SetRotateLog v1会按照24小时进行轮转切分，v2按照1GB进行文件切分  目前是hardcode的  暂时没有给配置接口
 
2. prometheus
 - 默认注册了requests_total  request_duration_millisecond response_size_bytes request_size_bytes
 
3. pprof
 - 参见相关路由


 
[具体使用，参考example文件](https://github.com/michael-kj/utils/blob/master/example/main.go)
