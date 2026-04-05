package llmstream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ChunkParser func(payload []byte) (chunk string, done bool, err error)

func Stream(client *http.Client, req *http.Request, onChunk func(chunk string) error, parser ChunkParser) error {
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stream API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}

		if line == "" {
			continue
		}

		if line == "[DONE]" {
			break
		}

		chunk, done, err := parser([]byte(line))
		if err != nil {
			return err
		}

		if chunk != "" && onChunk != nil {
			if err := onChunk(chunk); err != nil {
				return err
			}
		}

		if done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream read error: %w", err)
	}

	return nil
}

func ParseOpenAICompatibleChunk(payload []byte) (string, bool, error) {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(payload, &chunk); err != nil {
		return "", false, fmt.Errorf("failed to parse stream chunk: %w", err)
	}
	if len(chunk.Choices) == 0 {
		return "", false, nil
	}
	done := chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != ""
	return chunk.Choices[0].Delta.Content, done, nil
}

func ParseOllamaChunk(payload []byte) (string, bool, error) {
	var chunk struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Done bool `json:"done"`
	}
	if err := json.Unmarshal(payload, &chunk); err != nil {
		return "", false, fmt.Errorf("failed to parse stream chunk: %w", err)
	}
	return chunk.Message.Content, chunk.Done, nil
}
