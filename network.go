package main

import (
	"fmt"
	"strconv"
	"strings"
	_ "log"
	"github.com/godbus/dbus/v5"
)

type Network struct {
	Band  string `json:"band"`
	Arfcn string `json:"arfcn"`
	Pci   string `json:"pci"`
	Rsrp  string `json:"rsrp"`
	Rsrq  string `json:"rsrq"`
	Sinr  string `json:"sinr"`
}
type QoS struct {
    QCI int `json:"qci"`
    DL  int `json:"dl"`  // 下行带宽（Downlink）
    UL  int `json:"ul"`  // 上行带宽（Uplink）
}
// 辅助函数：分割字符串并Trim空格
func splitAndTrim(part []rune) []string {
	s := strings.TrimSpace(string(part))
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		res = append(res, strings.TrimSpace(p))
	}
 
	return res
}
func parseCellToVec(input string) [][]string {
	// 清理输入：去除首尾空格、结尾的OK、换行符
	cleaned := strings.TrimSpace(input)
	cleaned = strings.TrimSuffix(cleaned, "OK")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")

	runes := []rune(cleaned)
	var result [][]string
	var currentPart []rune
	var prevChar rune

	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if c == '-' {
			if prevChar == ',' {
				// 规则2：,- 作为负数处理，保留-
				currentPart = append(currentPart, c)
			} else {
				// 检查下一个字符是否是-
				if i+1 < len(runes) && runes[i+1] == '-' {
					// 规则3：遇到--，分割换行并保留第二个-
					if len(currentPart) > 0 {
						result = append(result, splitAndTrim(currentPart))
						currentPart = nil
					}
					currentPart = append(currentPart, '-')
					i++ // 跳过下一个-
				} else {
					// 规则1：单独-换行
					if len(currentPart) > 0 {
						result = append(result, splitAndTrim(currentPart))
						currentPart = nil
					}
				}
			}
		} else {
			currentPart = append(currentPart, c)
		}
		prevChar = c
	}

	// 处理最后剩余部分
	if len(currentPart) > 0 {
		result = append(result, splitAndTrim(currentPart))
	}

	// 打印 result
	// fmt.Println("Result:", result)

	return result
}

func parseOneCell(technology string, oneCell map[string]dbus.Variant, atResponse string) Network {
	oneData := parseCellToVec(atResponse)
	cell := Network{
		Band:  joinOrDefault(oneData, 0, ","),
		Arfcn: joinOrDefault(oneData, 1, ","),
		Pci: func() string {
			switch technology {
			case "nr":
				return getOrDefault(oneData, 2, 0)
			case "lte":
				return joinOrDefault(oneData, 2, ",")
			default:
				return ""
			}
		}(),
		Rsrp: formatValues(oneData, 3),
		Rsrq: formatValues(oneData, 4),
		Sinr: func() string {
			switch technology {
			case "nr":
				return formatValues(oneData, 15)
			case "lte":
				if val, ok := oneCell["ReferenceSignalSignalToNoiseRatio"].Value().(int32); ok {
					return fmt.Sprintf("%.2f", float64(val))
				}
				return "0.00"
			default:
				return "0.00"
			}
		}(),
	}
	return cell
}

func parseAllCell(technology string, atResponse string) []Network {
	allData := parseCellToVec(atResponse)
	networks := []Network{}

	switch technology {
	case "nr":
		if len(allData) > 0 {
			firstVec := allData[0]
			for i := 0; i < len(firstVec); i++ {
				networks = append(networks, Network{
					Band:  getOrDefault(allData, 0, i),
					Arfcn: getOrDefault(allData, 1, i),
					Pci:   getOrDefault(allData, 2, i),
					Rsrp:  formatValue(allData, 3, i),
					Rsrq:  formatValue(allData, 4, i),
					Sinr:  formatValue(allData, 5, i),
				})
			}
		}
	case "lte":
		for _, cellData := range allData {
			if len(cellData) < 13 || cellData[12] == "0" {
				break
			}
			networks = append(networks, Network{
				Arfcn: cellData[0],
				Pci:   cellData[1],
				Rsrp:  formatSingleValue(cellData[2]),
				Rsrq:  formatSingleValue(cellData[3]),
				Band:  cellData[12],
				Sinr:  "-",
			})
		}
	default:
	}

	return networks
}

func joinOrDefault(data [][]string, index int, sep string) string {
	if index < len(data) {
		return strings.Join(data[index], sep)
	}
	return ""
}

func getOrDefault(data [][]string, index, subIndex int) string {
	if index < len(data) && subIndex < len(data[index]) {
		return data[index][subIndex]
	}
	return ""
}

func formatValues(data [][]string, index int) string {
	if index >= len(data) {
		return ""
	}
	values := data[index]
	var formattedValues []string
	for _, val := range values {
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			formattedValues = append(formattedValues, fmt.Sprintf("%.2f", num/100.0))
		}
	}
	return strings.Join(formattedValues, ",")
}

func formatValue(data [][]string, index, subIndex int) string {
	if index < len(data) && subIndex < len(data[index]) {
		if num, err := strconv.ParseFloat(data[index][subIndex], 64); err == nil {
			return fmt.Sprintf("%.2f", num/100.0)
		}
	}
	return ""
}

func formatSingleValue(val string) string {
	if num, err := strconv.ParseFloat(val, 64); err == nil {
		return fmt.Sprintf("%.2f", num/100.0)
	}
	return ""
}
// 新增解析函数
func parseQoS(response string) QoS {
    cleaned := strings.TrimPrefix(response, "+CGEQOSRDP: ")
    cleaned = strings.ReplaceAll(cleaned, "\r", "")
    cleaned = strings.ReplaceAll(cleaned, "\n", "")
	cleaned = strings.TrimSuffix(cleaned, "OK")

    parts := strings.Split(strings.TrimSpace(cleaned), ",")
    
    getVal := func(index int) int {
        if len(parts) > index {
            val, _ := strconv.Atoi(strings.TrimSpace(parts[index]))
            return val
        }
        return 0
    }

    return QoS{
        QCI: getVal(1),
        DL:  getVal(6),
        UL:  getVal(7),
    }
}