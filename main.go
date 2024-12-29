package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/hajimehoshi/oto/v2"
	"github.com/hraban/opus"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v4"
)

// 创建一个循环缓冲区来处理音频数据
type AudioBuffer struct {
	buffer []byte
	mu     sync.Mutex
	cond   *sync.Cond
}

func NewAudioBuffer() *AudioBuffer {
	ab := &AudioBuffer{
		buffer: make([]byte, 0, 48000*2), // 1秒的音频数据
	}
	ab.cond = sync.NewCond(&ab.mu)
	return ab
}

func (ab *AudioBuffer) Write(data []byte) {
	ab.mu.Lock()
	ab.buffer = append(ab.buffer, data...)
	ab.cond.Signal()
	ab.mu.Unlock()
}

func (ab *AudioBuffer) Read(p []byte) (n int, err error) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	// 等待数据可用
	for len(ab.buffer) == 0 {
		ab.cond.Wait()
	}

	// 读取可用数据
	n = copy(p, ab.buffer)
	ab.buffer = ab.buffer[n:]
	return n, nil
}

type RealtimeClient struct {
	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel
	mediaTrack     mediadevices.Track
}

func NewRealtimeClient() (*RealtimeClient, error) {
	// WebRTC 配置
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// 创建 PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("创建 PeerConnection 失败: %v", err)
	}

	// 处理远端音频轨道
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			go handleRemoteTrack(track)
		}
	})

	client := &RealtimeClient{
		peerConnection: peerConnection,
	}

	return client, nil
}

// 处理远端音频轨道
func handleRemoteTrack(track *webrtc.TrackRemote) {
	// 创建 Opus 解码器
	decoder, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		fmt.Printf("创建 Opus 解码器失败: %v\n", err)
		return
	}

	// 初始化 oto 上下文
	ctx, ready, err := oto.NewContext(sampleRate, channels, 2)
	if err != nil {
		fmt.Printf("创建 oto 上下文失败: %v\n", err)
		return
	}
	<-ready

	// 创建音频缓冲区
	audioBuffer := NewAudioBuffer()

	// 创建播放器
	player := ctx.NewPlayer(audioBuffer)
	defer player.Close()

	// 开始播放
	player.Play()

	buffer := make([]int16, frameSize)
	pcm := make([]int16, frameSize)      // 解码后的 PCM 数据
	samples := make([]byte, frameSize*2) // 用于转换的临时缓冲区

	for {
		// 读取 RTP 包
		rtp, _, err := track.ReadRTP()
		if err != nil {
			fmt.Printf("读取 RTP 包失败: %v\n", err)
			continue
		}

		// 解码 Opus 数据
		n, err := decoder.Decode(rtp.Payload, pcm)
		if err != nil {
			fmt.Printf("解码音频失败: %v\n", err)
			continue
		}

		// 复制解码后的数据到播放缓冲区
		copy(buffer, pcm[:n])

		// 将 int16 数据转换为字节序列
		for i := 0; i < n; i++ {
			samples[i*2] = byte(buffer[i])
			samples[i*2+1] = byte(buffer[i] >> 8)
		}

		// 写入音频数据到缓冲区
		audioBuffer.Write(samples[:n*2])
	}
}

func main() {
	// 从环境变量获取 API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置 OPENAI_API_KEY 环境变量")
		return
	}

	// 获取临时令牌
	ephemeralToken, err := getEphemeralToken(apiKey)
	if err != nil {
		fmt.Printf("获取临时令牌失败: %v\n", err)
		return
	}

	// 创建 Realtime 客户端
	client, err := NewRealtimeClient()
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return
	}

	// 初始化音频
	err = client.initAudio()
	if err != nil {
		fmt.Printf("初始化音频失败: %v\n", err)
		return
	}

	// 连接到 OpenAI Realtime API
	err = client.connectToRealtimeAPI(ephemeralToken)
	if err != nil {
		fmt.Printf("连接到 Realtime API 失败: %v\n", err)
		return
	}

	// TODO: 实现与 OpenAI Realtime API 的连接逻辑

	// 保持程序运行
	select {}
}
