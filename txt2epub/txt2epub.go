package txt2epub

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run() {
	inputDir := "../books"
	outputDir := "../epub_output"
	// 建立輸出資料夾
	os.MkdirAll(outputDir, os.ModePerm)
	// 遍歷資料夾中的所有 txt 檔案
	filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("錯誤：", err)
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".txt") {
			processAndConvert(path, outputDir)
		}
		return nil
	})
}

func processAndConvert(inputPath string, outputDir string) {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		fmt.Println("無法開啟檔案：", inputPath)
		return
	}
	defer inputFile.Close()

	var processedLines []string
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue // 跳過空白行
		}
		// 替換 <, >, 空格
		trimmed = strings.ReplaceAll(trimmed, "<", "[")
		trimmed = strings.ReplaceAll(trimmed, ">", "]")
		trimmed = strings.ReplaceAll(trimmed, " ", "")
		// 每行後插入空行
		processedLines = append(processedLines, trimmed)
		processedLines = append(processedLines, "")
	}

	// 輸出處理後的 txt
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ".txt")
	processedPath := filepath.Join(outputDir, baseName+"_processed.txt")
	os.WriteFile(processedPath, []byte(strings.Join(processedLines, "\n")), 0644)

	// 使用 pandoc 轉成 epub
	outputEPUB := filepath.Join(outputDir, baseName+".epub")
	coverPath := filepath.Join(filepath.Dir(inputPath), baseName+".jpg")
	cmd := exec.Command("pandoc", processedPath, "-o", outputEPUB,
		"--metadata", "title="+baseName,
		"--metadata", "author=匿名",
		"--epub-cover-image="+coverPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("pandoc 轉檔失敗 %s: %s\n", baseName, stderr.String())
	} else {
		fmt.Printf("已轉換：%s → %s\n", processedPath, outputEPUB)
	}
}
