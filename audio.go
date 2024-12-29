package main

import (
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // 导入麦克风驱动
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v4"
)

const (
	sampleRate    = 48000
	channels      = 1
	frameSize     = 960
	opusFrameSize = 960  // 20ms @ 48kHz
	maxDataBytes  = 1000 // 足够大的缓冲区用于 Opus 编码数据
)

func (c *RealtimeClient) initAudio() error {

	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}
	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithAudioEncoders(&opusParams),
	)

	// 初始化音频设备
	audio, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {
			c.SampleRate = prop.Int(sampleRate)
			c.ChannelCount = prop.Int(channels)
			c.SampleSize = prop.Int(16) // 16位采样
		},
		Video: nil,
		Codec: codecSelector,
	})
	if err != nil {
		return fmt.Errorf("获取音频设备失败: %v", err)
	}

	// 获取音频轨道
	audioTracks := audio.GetTracks()
	if len(audioTracks) == 0 {
		return fmt.Errorf("没有找到音频轨道")
	}
	c.mediaTrack = audioTracks[0]

	// 将音频轨道添加到 PeerConnection
	_, err = c.peerConnection.AddTransceiverFromTrack(c.mediaTrack,
		webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendrecv,
		})

	if err != nil {
		return fmt.Errorf("添加音频轨道到 PeerConnection 失败: %v", err)
	}

	return nil
}

// 辅助函数：将 int16 切片转换为字节切片
func int16ToByte(data []int16) []byte {
	bytes := make([]byte, len(data)*2)
	for i, sample := range data {
		bytes[i*2] = byte(sample)
		bytes[i*2+1] = byte(sample >> 8)
	}
	return bytes
}

// 辅助函数：将字节切片转换为 int16 切片
func byteToInt16(data []byte) []int16 {
	samples := make([]int16, len(data)/2)
	for i := range samples {
		samples[i] = int16(data[i*2]) | int16(data[i*2+1])<<8
	}
	return samples
}
