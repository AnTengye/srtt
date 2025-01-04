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
	"github.com/AnTengye/srtt/api/deeplx"
	"github.com/spf13/cobra"
)

// support engine
const (
	deeplxEngine = "deeplx"
	baiduEngine  = "baidu"
)

var (
	sourceLang     string
	targetLang     string
	inputFilePath  string
	outputFilePath string
	contextLength  int
	contextOffset  int
	coffeeLength   int
	coffeeTime     int

	// api
	engine        string
	baseUrl       string
	debug         bool
	retry         int
	retryWaitTime int

	// baidu
	key    string
	secret string
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
	translateCmd.Flags().IntVarP(&contextLength, "ctxLength", "", 10, "Context length, too long may cause error")
	translateCmd.Flags().IntVarP(&contextOffset, "ctxOffset", "", 3, "Context offset")

	translateCmd.Flags().IntVarP(&coffeeLength, "coffeeLength", "", 400, "Drink coffee when processing {n} lines.0 means no coffee")
	translateCmd.Flags().IntVarP(&coffeeTime, "coffeeTime", "", 500, "Drink coffee for {n} ms")

	translateCmd.Flags().StringVarP(&engine, "engine", "e", "deeplx", "Translation engine: deeplx, baidu")
	translateCmd.Flags().StringVarP(&baseUrl, "apiUrl", "", "", "API base url")
	translateCmd.Flags().BoolVarP(&debug, "debug", "", false, "Debug mode")
	translateCmd.Flags().IntVarP(&retry, "retry", "", 3, "Retry times")
	translateCmd.Flags().IntVarP(&retryWaitTime, "retryWT", "", 500, "Retry wait time(ms)")

	translateCmd.Flags().StringVarP(&key, "apiKey", "", "", "Api key, required for some engine")
	translateCmd.Flags().StringVarP(&secret, "apiSecret", "", "", "Api secret, required for some engine")
}

func translateRun(cmd *cobra.Command, args []string) {
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
		apiClient = deeplx.NewClient(logger.With("engine", deeplxEngine),
			api.WithBaseUrl(baseUrl),
			api.WithDebug(debug),
			api.WithRetry(retry),
			api.WithRetryWaitTime(time.Duration(retryWaitTime)*time.Millisecond),
		)
	case baiduEngine:
		if key == "" || secret == "" {
			logger.Fatal("baidu api key or secret is empty")
		}
		apiClient = baidu.NewClient(key, secret, logger.With("engine", baiduEngine),
			api.WithBaseUrl(baseUrl),
			api.WithDebug(debug),
			api.WithRetry(retry),
			api.WithRetryWaitTime(time.Duration(retryWaitTime)*time.Millisecond),
		)
	default:
		logger.Fatal("暂时不支持的翻译引擎")
	}
	translatedText := processText(apiClient, srtLines, contextLength, contextOffset)
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
	logger.Infof("翻译完成，结果已保存到 %s", outputFilePath)
}

func processText(client api.TranslateApi, lines []string, blockSize, overlap int) []string {
	translatedText := make([]string, len(lines))
	for i := 0; i < len(lines); i += blockSize - overlap {
		if coffeeLength != 0 && i%coffeeLength == 0 {
			time.Sleep(time.Duration(coffeeTime) * time.Millisecond)
			logger.Infof("已翻译%d行，累了，喝杯咖啡", i)
		}
		// 确保不会超出切片范围
		end := i + blockSize
		if end > len(lines) {
			end = len(lines)
		}

		// 提取当前窗口的行
		currentBlock := lines[i:end]

		logger.Debugf("原文：\n %s", strings.Join(currentBlock, "\n"))
		sourceText := strings.Join(currentBlock, "\n----\n")
		// 调用处理函数
		result, err := client.Translate(sourceText, sourceLang, targetLang)
		if err != nil {
			logger.Errorf("第%d-%d行翻译失败", i+1, end+1)
			continue
		}
		split := strings.Split(result, "----")
		logger.Debugf("译文：\n%s", split)
		if i > 0 {
			// 如果不是第一轮，则只选择blockSize-overlap个翻译结果
			for j := overlap; j < len(currentBlock) && j < len(split); j++ {
				translatedText[i+j] = strings.TrimSpace(split[j])
			}
		} else {
			for j := 0; j < len(currentBlock) && j < len(split); j++ {
				translatedText[i+j] = strings.TrimSpace(split[j])
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
