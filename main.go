package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// 发送at命令请求结构体
type RequestPayload struct {
	AtCmd string `json:"atcmd"`
	Path  string `json:"path"`
}

// 通用响应结构体
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// at命令api
func sendAtCmdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	var requestPayload RequestPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestPayload); err != nil {
		fmt.Fprintln(w, "无效请求参数:", err.Error())
		return
	}
	response, err := sendAtCmd(requestPayload.AtCmd, requestPayload.Path)
	if err != nil {
		fmt.Fprintln(w, "", err.Error())
		return
	}
	fmt.Fprintln(w, response)
}

// 新增短信发送接口
func sendSMSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Path  string `json:"path"`
		Phone string `json:"to"`
		Text  string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(Response{
			Code: http.StatusBadRequest,
			Msg:  "无效请求参数: " + err.Error(),
			Data: nil,
		})
		return
	}

	resultPath, err := sendSMS(req.Path, req.Phone, req.Text)
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Code: http.StatusOK,
		Msg:  "短信发送成功",
		Data: map[string]string{"path": resultPath},
	})
}

func sysInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sysInfo, err := getSysInfo()
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "" + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	response := Response{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: sysInfo,
	}
	json.NewEncoder(w).Encode(response)
}

type QueryRequestPayload struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func querySMSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var requestPayload QueryRequestPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestPayload); err != nil {
		response := Response{
			Code: http.StatusBadRequest,
			Msg:  "请求参数无效: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	if requestPayload.Page <= 0 {
		requestPayload.Page = 1
	}
	if requestPayload.Limit <= 0 {
		requestPayload.Limit = 10
	}
	initDB(defaultDbPath)
	db, err := sql.Open("sqlite", defaultDbPath)
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "打开数据库失败: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer db.Close()

	queryResponse, err := querySMS(db, requestPayload.Page, requestPayload.Limit)
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "查询短信错误: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := Response{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: queryResponse,
	}
	json.NewEncoder(w).Encode(response)
}

type DeleteRequestPayload struct {
	IDs []int `json:"ids"`
}

func deleteSMSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var requestPayload DeleteRequestPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestPayload); err != nil {
		response := Response{
			Code: http.StatusBadRequest,
			Msg:  "请求参数无效: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	initDB(defaultDbPath)
	db, err := sql.Open("sqlite", defaultDbPath)
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "打开数据库失败: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer db.Close()
	rowsAffected, err := deleteSMS(db, requestPayload.IDs)
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "删除短信失败: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("成功删除 %d 条短信", rowsAffected),
		Data: nil,
	}
	json.NewEncoder(w).Encode(response)
}

type NetPayload struct {
	Path string `json:"path"`
}

func netInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var requestPayload NetPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestPayload); err != nil {
		response := Response{
			Code: http.StatusBadRequest,
			Msg:  "请求参数无效: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	cellInfo, err := getServingCellInformation(requestPayload.Path)
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "获取服务小区信息失败: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	technology, ok := cellInfo["Technology"]
	if !ok {
		response := Response{
			Code: http.StatusNotFound,
			Msg:  "未找到 Technology 信息",
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	technologyStr, ok := technology.Value().(string)
	if !ok {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "Technology 获取错误",
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	var oneCellAtCmd, allCellAtCmd string
	switch technologyStr {
	case "nr":
		oneCellAtCmd = "AT+SPENGMD=0,14,1"
		allCellAtCmd = "AT+SPENGMD=0,14,2"
	case "lte":
		oneCellAtCmd = "AT+SPENGMD=0,6,0"
		allCellAtCmd = "AT+SPENGMD=0,6,6"
	default:
		oneCellAtCmd = "AT+SPENGMD=0,14,1"
		allCellAtCmd = "AT+SPENGMD=0,14,2"
	}
	oneCellResp, err := sendAtCmd(oneCellAtCmd, requestPayload.Path)
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "获取小区失败: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	allCellResp, err := sendAtCmd(allCellAtCmd, requestPayload.Path)
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "获取邻区失败: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	qosResp, err := sendAtCmd("AT+CGEQOSRDP=1", requestPayload.Path)
	if err != nil {
		response := Response{
			Code: http.StatusInternalServerError,
			Msg:  "获取QOS失败: " + err.Error(),
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	oneCellVec := parseOneCell(technologyStr, cellInfo, oneCellResp)
	allCells := parseAllCell(technologyStr, allCellResp)
	qosInfo := parseQoS(qosResp)
	networks := []Network{oneCellVec}
	networks = append(networks, allCells...)

	response := Response{
		Code: http.StatusOK,
		Msg:  "Success",
		Data: struct {
			Tech     string    `json:"tech"`
			Networks []Network `json:"cells"`
			QoS      QoS       `json:"qos"`
		}{
			Tech:     technologyStr,
			Networks: networks,
			QoS:      qosInfo,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// 跨域中间件,处理响应头和OPTIONS请求
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

const defaultHost string = "0.0.0.0"
const defaultPort int = 8000
const defaultDbPath string = "./sms.db"

func main() {

	initDB(defaultDbPath)
	go startSMSListener()

	// 注册路由
	http.Handle("/", corsMiddleware(http.FileServer(http.Dir("./www"))))
	//at命令
	http.Handle("/api/at", corsMiddleware(http.HandlerFunc(sendAtCmdHandler)))
	//短信发送
	http.Handle("/api/sms/send", corsMiddleware(http.HandlerFunc(sendSMSHandler)))
	//分页查询
	http.Handle("/api/sms/query", corsMiddleware(http.HandlerFunc(querySMSHandler)))
	//批量删除
	http.Handle("/api/sms/delete", corsMiddleware(http.HandlerFunc(deleteSMSHandler)))
	//基站信息，qos，邻区
	http.Handle("/api/network", corsMiddleware(http.HandlerFunc(netInfoHandler)))
	//系统信息
	http.Handle("/api/sysinfo", corsMiddleware(http.HandlerFunc(sysInfoHandler)))
	// 监听端口
	addr := fmt.Sprintf("%s:%d", defaultHost, defaultPort)
	log.Printf("服务器正在监听 %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
