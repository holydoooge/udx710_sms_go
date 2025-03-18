package main

import (
	"database/sql"
	_ "fmt"
	"log"

	"github.com/godbus/dbus/v5"
)

func sendAtCmd(atCmd, path string) (string, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	obj := conn.Object("org.ofono", dbus.ObjectPath(path))
	var response string
	err = obj.Call("org.ofono.Modem.SendAtcmd", 0, atCmd).Store(&response)
	if err != nil {
		return "", err
	}
	return response, nil
}

func getServingCellInformation(path string) (map[string]dbus.Variant, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	obj := conn.Object("org.ofono", dbus.ObjectPath(path))
	var cellInfo map[string]dbus.Variant
	err = obj.Call("org.ofono.NetworkMonitor.GetServingCellInformation", 0).Store(&cellInfo)
	if err != nil {
		return nil, err
	}

	return cellInfo, nil
}
func sendSMS(path, to, text string) (string, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	obj := conn.Object("org.ofono", dbus.ObjectPath(path))
	var resultPath dbus.ObjectPath
	err = obj.Call("org.ofono.MessageManager.SendMessage", 0, to, text).Store(&resultPath)
	if err != nil {
		return "", err
	}
	return string(resultPath), nil
}

func startSMSListener() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Printf("D-Bus连接失败: %v", err)
		return
	}
	defer conn.Close()

	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.ofono.MessageManager"),
		dbus.WithMatchMember("IncomingMessage"),
	); err != nil {
		log.Printf("添加信号匹配失败: %v", err)
		return
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)

	for v := range c {
		if v.Name == "org.ofono.MessageManager.IncomingMessage" && len(v.Body) >= 2 {
			content := v.Body[0].(string)
			infoDict := v.Body[1].(map[string]dbus.Variant)
			sender := getVariantString(infoDict["Sender"])
			sentTime := getVariantString(infoDict["SentTime"])
			localTime := getVariantString(infoDict["LocalSentTime"])
			initDB(defaultDbPath)
			db, err := sql.Open("sqlite", defaultDbPath)
			if err != nil {
				log.Printf("打开数据库失败: %v", err)
				continue
			}
			rowsAffected, err := insertSMS(db, sender, content, localTime, sentTime)
			if err != nil {
				log.Printf("插入短信失败: %v", err)
				db.Close()
				continue
			}
			log.Printf("插入短信成功，受影响行数: %d", rowsAffected)
			db.Close()
		} else {
			log.Println("收到无效的短信信号格式")
		}
	}
}

func getVariantString(v dbus.Variant) string {
	if s, ok := v.Value().(string); ok {
		return s
	}
	return ""
}
