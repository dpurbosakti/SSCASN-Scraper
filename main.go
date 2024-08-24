package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	baseURL     = "https://api-sscasn.bkn.go.id/2024/portal/spf"
	kodeRefPend = "4609080"
	namaJurusan = "DIII Fisioterapi"
)

var headers = map[string]string{
	"accept":             "application/json, text/plain, */*",
	"accept-encoding":    "gzip, deflate, br, zstd",
	"accept-language":    "en-US,en;q=0.9,id-ID;q=0.8,id;q=0.7",
	"connection":         "keep-alive",
	"host":               "api-sscasn.bkn.go.id",
	"origin":             "https://sscasn.bkn.go.id",
	"referer":            "https://sscasn.bkn.go.id/",
	"sec-ch-ua":          "\"Not)A;Brand\";v=\"99\", \"Google Chrome\";v=\"127\", \"Chromium\";v=\"127\"",
	"sec-ch-ua-mobile":   "?1",
	"sec-ch-ua-platform": "\"Android\"",
	"sec-fetch-dest":     "empty",
	"sec-fetch-mode":     "cors",
	"sec-fetch-site":     "same-site",
	"user-agent":         "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Mobile Safari/537.36",
}

type Response struct {
	Data struct {
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
		Data []map[string]interface{} `json:"data"`
	} `json:"data"`
}

func setNamaJurusan(namaJurusan string) string {
	return strings.ReplaceAll(namaJurusan, " ", "_")
}

func fetchData(offset int, retries int, delay time.Duration) (*Response, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s?kode_ref_pend=%s&offset=%d", baseURL, kodeRefPend, offset)

	for i := 0; i < retries; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case 200:
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var response Response
			err = json.Unmarshal(body, &response)
			if err != nil {
				fmt.Printf("Error decoding JSON at offset %d: Response is not valid JSON.\n", offset)
				return nil, err
			}

			return &response, nil
		case 504:
			fmt.Printf("Request failed at offset %d: 504 Gateway Timeout. Retry %d/%d...\n", offset, i+1, retries)
			time.Sleep(delay)
		default:
			fmt.Printf("Request failed at offset %d: %d\n", offset, resp.StatusCode)
			return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
		}
	}

	return nil, fmt.Errorf("failed to fetch data after %d retries", retries)
}

func main() {
	fmt.Println("Memulai proses pengambilan data...")

	initialData, err := fetchData(0, 3, 5*time.Second)
	if err != nil {
		log.Fatal("Gagal mengambil data awal:", err)
	}

	totalData := initialData.Data.Meta.Total
	fmt.Printf("Total data ditemukan: %d\n", totalData)

	timestamp := time.Now().Format("20060102_150405")
	dataDir := "data"
	excelOutputFile := filepath.Join(dataDir, fmt.Sprintf("sscasn_data_%s.xlsx", setNamaJurusan(namaJurusan)+"_"+timestamp))

	// Create the data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Error creating data directory:", err)
	}

	var allData []map[string]interface{}

	for offset := 0; offset < totalData; offset += 10 {
		fmt.Printf("Mengambil data dengan offset %d...\n", offset)
		data, err := fetchData(offset, 3, 5*time.Second)
		if err != nil {
			fmt.Printf("Error fetching data at offset %d: %v\n", offset, err)
			continue
		}
		allData = append(allData, data.Data.Data...)
	}

	fmt.Println("Membuat file Excel...")

	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	// Set headers
	headers := []string{"ins_nm", "jp_nm", "formasi_nm", "jabatan_nm", "lokasi_nm", "jumlah_formasi", "gaji_min", "gaji_max", "pengumuman"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue(sheet, cell, header)
	}

	// Set metadata
	f.SetCellValue(sheet, "A1", "updated_at")
	f.SetCellValue(sheet, "B1", time.Now().Format("2006-01-02 15:04:05"))
	f.SetCellValue(sheet, "A2", "auto_update_by")
	f.SetCellValue(sheet, "B2", "moko")

	// Set data
	for i, record := range allData {
		gajiMin, _ := strconv.ParseFloat(record["gaji_min"].(string), 64)
		gajiMax, _ := strconv.ParseFloat(record["gaji_max"].(string), 64)

		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+5), record["ins_nm"])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+5), record["jp_nama"])
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+5), record["formasi_nm"])
		f.SetCellValue(sheet, fmt.Sprintf("D%d", i+5), record["jabatan_nm"])
		f.SetCellValue(sheet, fmt.Sprintf("E%d", i+5), record["lokasi_nm"])
		f.SetCellValue(sheet, fmt.Sprintf("F%d", i+5), record["jumlah_formasi"])
		f.SetCellValue(sheet, fmt.Sprintf("G%d", i+5), gajiMin)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", i+5), gajiMax)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", i+5), fmt.Sprintf("https://sscasn.bkn.go.id/detailformasi/%v", record["formasi_id"]))
	}

	// Auto-fit column width
	for i := 1; i <= 8; i++ {
		col, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, col, col, 30) // Increased width for better readability
	}

	if err := f.SaveAs(excelOutputFile); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Proses selesai! Data berhasil disimpan dalam file %s\n", excelOutputFile)
}
