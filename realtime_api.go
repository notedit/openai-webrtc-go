package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pion/webrtc/v4"
)

type SessionResponse struct {
	ClientSecret struct {
		Value string `json:"value"`
	} `json:"client_secret"`
}

func (c *RealtimeClient) connectToRealtimeAPI(ephemeralToken string) error {
	offer, err := c.peerConnection.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("创建 offer 失败: %v", err)
	}

	err = c.peerConnection.SetLocalDescription(offer)
	if err != nil {
		return fmt.Errorf("设置本地描述失败: %v", err)
	}

	baseURL := "https://api.openai.com/v1/realtime"
	model := "gpt-4o-realtime-preview-2024-12-17"

	req, err := http.NewRequest("POST", fmt.Sprintf("%s?model=%s", baseURL, model), bytes.NewReader([]byte(offer.SDP)))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+ephemeralToken)
	req.Header.Set("Content-Type", "application/sdp")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	answerSDP, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	err = c.peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  string(answerSDP),
	})
	if err != nil {
		return fmt.Errorf("设置远程描述失败: %v", err)
	}

	return nil
}

func getEphemeralToken(apiKey string) (string, error) {
	payload := map[string]interface{}{
		"model": "gpt-4o-realtime-preview-2024-12-17",
		"voice": "verse",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/realtime/sessions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var sessionResp SessionResponse
	err = json.Unmarshal(body, &sessionResp)
	if err != nil {
		return "", err
	}

	return sessionResp.ClientSecret.Value, nil
}
