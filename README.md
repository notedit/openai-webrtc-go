# openai-webrtc-go

使用 Go 语言实现的 OpenAI Realtime API WebRTC 客户端，支持实时语音交互。

## 功能

- 基于 WebRTC 的实时音频通信
- 本地音频采集（使用 Pion MediaDevices）
- 远程音频播放（使用 Oto）
- Opus 音频编解码
- 支持 OpenAI 临时令牌认证

## 快速开始

### 前置要求

- Go 1.21+
- OpenAI API Key

### 安装

```bash
git clone https://github.com/yourusername/openai-webrtc-go.git
cd openai-webrtc-go
go mod tidy
```


### 运行

1. 设置环境变量：

```bash
export OPENAI_API_KEY=your_api_key
```

2. 运行程序：

```bash
go run .
```

## 技术栈

- [Pion WebRTC](https://github.com/pion/webrtc): WebRTC 实现
- [Pion MediaDevices](https://github.com/pion/mediadevices): 音频设备管理
- [Opus](https://github.com/hraban/opus): 音频编解码
- [Oto](https://github.com/hajimehoshi/oto): 音频播放

## License

MIT

## 相关链接

- [OpenAI Realtime API 文档](https://platform.openai.com/docs/api-reference/realtime)