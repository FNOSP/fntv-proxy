# fnTv-vlcProxy

## 📖 简介
　　这是一个用Go语言编写的第三方飞牛影视客户端fntv-client-multiplatform项目的VLC播放器代理桥接程序，程序充当代理角色，同时兼容飞牛影视的HTTP及HTTPS连接请求，用于解决调用VLC播放器连接飞牛影视服务器时不能传递如cookie等自定义请求头的问题。

　　第三方跨平台飞牛影视客户端项目地址：[fntv-client-multiplatform](https://github.com/FNOSP/fntv-client-multiplatform)

## ✨ 功能特点

- 🌐 在本地启动HTTP服务接收来自VLC的请求
- 🍪 支持设置自定义Cookie请求头
- 📊 自动处理特定路径的Range请求头
- 🔗 兼容HTTP和HTTPS连接请求
- ⚙️ 通过配置文件动态管理服务端口

## ⚙️ 配置文件

程序会在同目录下查找`config.ini`文件，格式如下：

```ini
[server]
port=1999
```

如果配置文件不存在，程序会自动创建一个默认配置文件，端口为1999。

## 🚀 使用方法

### 🏃 启动服务

```bash
go run main.go
```

或编译后运行：

```bash
go build -o vlc-proxy
./vlc-proxy
```

### 🔧 设置代理信息

发送POST请求到`/proxyInfo`接口，参数：
- `url`: 目标服务器地址（如：http://192.168.1.200:5666）
- `cookie`: Cookie信息（如：Trim-MC-token=2a075b3438764b4da9e772c66a759548; lastLoginUsername=admin）

示例：
```bash
curl -X POST -d "url=http://192.168.1.200:5666&cookie=Trim-MC-token=2a075b3438764b4da9e772c66a759548; lastLoginUsername=admin" http://127.0.0.1:1999/proxyInfo
```

### 📺 VLC播放设置

在VLC中设置播放地址为本地代理服务地址，例如：
```
http://127.0.0.1:1999/v/media/ac611442ed3fb17daa73e71bc1268d02/preset.m3u8
```

程序会自动将请求转发到目标服务器：
```
http://192.168.1.200:5666/v/media/ac611442ed3fb17daa73e71bc1268d02/preset.m3u8
```

并添加预设的Cookie信息。

## ⚠️ 注意事项

- 本程序仅用于适配飞牛影视服务的VLC播放器代理，其他影视服务器不支持！
- 确保目标服务器地址正确且可访问
- Cookie信息需要完整，包括所有必要的键值对
- 程序会保留原始请求的大部分HTTP头部信息（除Host外）
