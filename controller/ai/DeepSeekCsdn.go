package ai

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"
)

// 定义响应结构
type ChatResponse struct {
    Choices []struct {
        Delta struct {
            Content string `json:"content"`
        } `json:"delta"`
    } `json:"choices"`
}

func main() {
    // 创建输出文件
    file, err := os.OpenFile("conversation.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return
    }
    defer file.Close()

    // API 配置
    url := "https://api.siliconflow.cn/v1/chat/completions"
    
    for {
        // 获取用户输入
        fmt.Print("\n请输入您的问题 (输入 q 退出): ")
        reader := bufio.NewReader(os.Stdin)
        question, _ := reader.ReadString('\n')
        question = strings.TrimSpace(question)
        
        if question == "q" {
            break
        }

        // 记录对话时间
        timestamp := time.Now().Format("2006-01-02 15:04:05")
        file.WriteString(fmt.Sprintf("\n[%s] Question:\n%s\n\n", timestamp, question))

        // 构建请求体
        payload := fmt.Sprintf(`{
            "model": "deepseek-ai/DeepSeek-V3",
            "messages": [
                {
                    "role": "user",
                    "content": "%s"
                }
            ],
            "stream": true,
            "max_tokens": 2048,
            "temperature": 0.7
        }`, question)

        // 发送请求
        req, _ := http.NewRequest("POST", url, strings.NewReader(payload))
        req.Header.Add("Content-Type", "application/json")
        req.Header.Add("Authorization", "Bearer YOUR_API_KEY")  // 替换为你的 API Key

        // 获取响应
        res, _ := http.DefaultClient.Do(req)
        defer res.Body.Close()

        // 处理流式响应
        scanner := bufio.NewReader(res.Body)
        for {
            line, err := scanner.ReadString('\n')
            if err != nil {
                break
            }

            line = strings.TrimSpace(line)
            if line == "" || line == "data: [DONE]" {
                continue
            }

            if strings.HasPrefix(line, "data: ") {
                line = strings.TrimPrefix(line, "data: ")
            }

            var response ChatResponse
            if err := json.Unmarshal([]byte(line), &response); err != nil {
                continue
            }

            if len(response.Choices) > 0 {
                content := response.Choices[0].Delta.Content
                if content != "" {
                    fmt.Print(content)
                    file.WriteString(content)
                }
            }
        }
    }
}
