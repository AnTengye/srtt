/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AnTengye/srtt/api"
	"github.com/AnTengye/srtt/api/baidu"
	"github.com/AnTengye/srtt/api/chatgpt"
	"github.com/AnTengye/srtt/api/deeplx"
	"github.com/AnTengye/srtt/api/google"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

// support engine
const (
	deeplxEngine  = "deeplx"
	baiduEngine   = "baidu"
	chatgptEngine = "chatgpt"
	googleEngine  = "google"
)

var (
	sourceLang     string
	targetLang     string
	inputFilePath  string
	outputFilePath string
	processLength  int
	contextOffset  int
	coffeeLength   int
	coffeeTime     int

	// api
	engine        string
	baseUrl       string
	retry         int
	retryWaitTime int

	key      string
	secret   string
	gptModel string
)

// translateCmd represents the translation command
var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "translate srt",
	Long: `
translate srt file to other language
`,
	Run: translateRun,
}

func init() {
	rootCmd.AddCommand(translateCmd)

	translateCmd.Flags().StringVarP(&sourceLang, "source", "s", "ja", "Source language")
	translateCmd.Flags().StringVarP(&targetLang, "target", "t", "zh", "Target language")
	translateCmd.Flags().StringVarP(&inputFilePath, "input", "i", "ja.srt", "Input file path")
	translateCmd.Flags().StringVarP(&outputFilePath, "output", "o", "", "Output file path")
	translateCmd.Flags().IntVarP(&processLength, "processLength", "", 10, "Length of each process")
	translateCmd.Flags().IntVarP(&contextOffset, "ctxOffset", "", 3, "Context offset")

	translateCmd.Flags().IntVarP(&coffeeLength, "coffeeLength", "", 0, "Drink coffee times. 0 means no coffee")
	translateCmd.Flags().IntVarP(&coffeeTime, "coffeeTime", "", 0, "Drink coffee need time, you should set coffeeLength. 0 means no coffee")

	translateCmd.Flags().StringVarP(&engine, "engine", "e", "deeplx", "Translation engine: deeplx, baidu")
	translateCmd.Flags().StringVarP(&baseUrl, "apiUrl", "", "", "API base url")
	translateCmd.Flags().IntVarP(&retry, "retry", "", 3, "Retry times")
	translateCmd.Flags().IntVarP(&retryWaitTime, "retryWT", "", 500, "Retry wait time(ms)")

	translateCmd.Flags().StringVarP(&key, "apiKey", "", "", "Api key, required for some engine")
	translateCmd.Flags().StringVarP(&secret, "apiSecret", "", "", "Api secret, required for some engine")
	translateCmd.Flags().StringVarP(&gptModel, "gptModel", "", "gpt-3.5-turbo", "GPT model")
}

func translateRun(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	defer func() {
		logger.Infof("耗时： %d s\n", int(time.Since(startTime).Seconds()))
	}()
	file, err := os.Open(inputFilePath)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer file.Close()
	var lines []string
	var srtLines []string
	scanner := bufio.NewScanner(file)
	srtIndex := 0
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if srtIndex%4 == 2 {
			srtLines = append(srtLines, scanner.Text())
		}
		srtIndex++
	}
	logger.Infof("文本行数: %d", len(srtLines))
	var apiClient api.TranslateApi
	switch engine {
	case deeplxEngine:
		apiClient = deeplx.NewClient(logger.With("engine", engine),
			deeplx.WithBaseUrl(baseUrl),
			deeplx.WithDebug(debug),
			deeplx.WithRetry(retry),
			deeplx.WithRetryWaitTime(time.Duration(retryWaitTime)*time.Millisecond),
		)
	case baiduEngine:
		if key == "" || secret == "" {
			logger.Fatal("baidu api key or secret is empty")
		}
		apiClient = baidu.NewClient(key, secret, logger.With("engine", engine),
			baidu.WithBaseUrl(baseUrl),
			baidu.WithDebug(debug),
			baidu.WithRetry(retry),
			baidu.WithRetryWaitTime(time.Duration(retryWaitTime)*time.Millisecond),
		)
	case chatgptEngine:
		apiClient = chatgpt.NewClient(key, logger.With("engine", engine),
			chatgpt.WithBaseUrl(baseUrl),
			chatgpt.WithModel(gptModel),
			chatgpt.WithCtxOffset(contextOffset),
		)
	case googleEngine:
		apiClient = google.NewClient(key, secret, logger.With("engine", engine), true)
	default:
		logger.Fatal("暂时不支持的翻译引擎")
	}
	defer apiClient.Close()
	if coffeeLength != 0 && coffeeTime != 0 {
		logger.Infof("已进入咖啡厅（速率限制模式）")
	}
	logger.Infof("开始翻译,当前引擎: %s", engine)
	translatedText := processText(apiClient, srtLines, processLength, contextOffset)
	// 将翻译后的文本写入文件
	if outputFilePath == "" {
		outputFilePath = formatFileNameWithoutExtension(inputFilePath, targetLang)
	}
	translatedFile, err := os.Create(outputFilePath)
	if err != nil {
		logger.Fatal(err)
	}
	defer translatedFile.Close()
	for i := 0; i < len(lines); i += 1 {
		if i%4 == 2 {
			_, err := translatedFile.WriteString(translatedText[i/4] + "\n")
			if err != nil {
				logger.Fatal(err)
			}
		} else {
			_, err := translatedFile.WriteString(lines[i] + "\n")
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

	logger.Infof("翻译完成, 结果已保存到 %s", outputFilePath)
}

func processText(client api.TranslateApi, lines []string, blockSize, overlap int) []string {
	translatedText := make([]string, len(lines))
	var throttler *rate.Limiter
	if coffeeLength != 0 && coffeeTime != 0 {
		throttler = rate.NewLimiter(
			rate.Every(time.Duration(coffeeTime)*time.Second),
			coffeeLength,
		)
	}
	for i := 0; i < len(lines); i += blockSize - overlap {
		if throttler != nil {
			reserve := throttler.Reserve()
			if reserve.Delay() != 0 {
				logger.Infof("已翻译%d行，喝杯咖啡需花费%d秒", i, int(reserve.Delay().Seconds()))
				time.Sleep(reserve.Delay())
			}
		}
		// 确保不会超出切片范围
		end := i + blockSize
		if end > len(lines) {
			end = len(lines)
		}

		// 提取当前窗口的行
		currentBlock := lines[i:end]
		logger.Infof("正在翻译第%d-%d行", i+1, end)
		logger.Debugf("原文：\n %s", strings.Join(currentBlock, "\n----\n"))
		// 调用处理函数
		result, err := client.Translate(currentBlock, sourceLang, targetLang)
		if err != nil {
			logger.Errorf("第%d-%d行翻译失败: %s", i+1, end, err)
			continue
		}
		logger.Debugf("译文：\n%s", result)
		if i > 0 {
			// 如果不是第一轮，则只选择blockSize-overlap个翻译结果
			for j := overlap; j < len(currentBlock) && j < len(result); j++ {
				translatedText[i+j] = strings.TrimSpace(result[j])
			}
		} else {
			for j := 0; j < len(currentBlock) && j < len(result); j++ {
				translatedText[i+j] = strings.TrimSpace(result[j])
			}
		}
	}
	return translatedText
}

func formatFileNameWithoutExtension(fileName, targetLang string) string {
	lastDotIndex := strings.LastIndex(fileName, ".")
	if lastDotIndex > -1 {
		fileName = fileName[:lastDotIndex]
	}
	return fmt.Sprintf("%s_%s.srt", fileName, targetLang)
}
