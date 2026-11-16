// ตัวอย่างการใช้ log/slog: handler, level, attribute, group, LogValuer
package main

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// Password เป็นข้อมูลลับ — implement slog.LogValuer เพื่อไม่ให้ค่าจริงหลุดลง log
type Password string

func (Password) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

// User implement LogValuer เพื่อเลือกเฉพาะ field ที่ควร log
type User struct {
	ID       int
	Email    string
	Password Password
}

func (u User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("id", u.ID),
		slog.String("email", u.Email),
	)
}

// removeTime ลบ field time ออก เพื่อให้ output คงที่ (ใช้เฉพาะในตัวอย่าง)
func removeTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

func main() {
	// 1) TextHandler: อ่านง่ายสำหรับคน เหมาะกับตอน dev
	text := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:       slog.LevelDebug, // แสดงตั้งแต่ Debug ขึ้นไป
		ReplaceAttr: removeTime,
	}))
	text.Debug("กำลังโหลด config", "path", "config.yaml")
	text.Info("server started", "port", 8080)

	// 2) JSONHandler: เหมาะกับ production ให้ระบบเก็บ log parse ต่อได้
	opts := &slog.HandlerOptions{Level: slog.LevelInfo, ReplaceAttr: removeTime}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	logger.Debug("ไม่แสดง เพราะต่ำกว่า Info")
	logger.Warn("disk usage high", slog.Int("percent", 91))

	// 3) slog.With: แนบ attribute ที่ใช้ซ้ำทุกบรรทัด
	reqLog := logger.With("request_id", "req-42", "method", "GET")
	reqLog.Info("handling request", "path", "/users")

	// 4) Group: จัด attribute เป็นกลุ่ม
	logger.Info("db query",
		slog.Group("db", slog.String("table", "users"), slog.Duration("took", 150*time.Millisecond)),
	)

	// 5) LogValuer: ซ่อนข้อมูลลับ
	u := User{ID: 7, Email: "a@example.com", Password: "s3cret"}
	logger.Info("user login", "user", u, "password", u.Password)

	// 6) Context: ส่ง ctx ไปให้ handler (เช่น ดึง trace id) และ LogAttrs ที่เร็วกว่า
	ctx := context.Background()
	logger.LogAttrs(ctx, slog.LevelError, "payment failed",
		slog.String("order_id", "o-1001"),
		slog.Any("err", os.ErrDeadlineExceeded),
	)

	// 7) ตั้งเป็น default: slog.Info และ log.Printf จะใช้ handler นี้ด้วย
	slog.SetDefault(logger)
	slog.Info("ใช้ default logger แล้ว")
}
