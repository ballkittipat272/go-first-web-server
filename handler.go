package main

import (
	"fmt"
	"net/http"
	"sync"
)

// CounterHandler เป็นตัวอย่างของ "Stateful Handler"
// ที่เก็บ state (counter) และจัดการ concurrency ด้วย mutex
type CounterHandler struct {
	mu      sync.Mutex
	counter int
}

// ServeHTTP ทำให้ CounterHandler implement http.Handler interface
// ทุกครั้งที่ endpoint นี้ถูกเรียก, counter จะเพิ่มขึ้นอย่างปลอดภัย (thread-safe)
func (h *CounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Lock เพื่อป้องกัน race condition จาก goroutine อื่นๆ
	h.mu.Lock()
	h.counter++
	// อ่านค่า counter ไปเก็บในตัวแปร local ก่อนที่จะ unlock
	// เพื่อให้แน่ใจว่าค่าที่แสดงผลเป็นค่าที่ถูกต้อง ณ เวลาที่อ่าน
	count := h.counter
	h.mu.Unlock()

	fmt.Fprintf(w, "This endpoint was called %d times\n", count)
}

func main() {
	// สร้าง instance ของ handler ขึ้นมาเพียงครั้งเดียว
	// state ของ handler (counter) จะถูกแชร์ระหว่างทุกๆ request ที่เข้ามา
	handler := &CounterHandler{}
	http.Handle("/count", handler)

	fmt.Println("Server is listening on :8080")
	fmt.Println("Try accessing http://localhost:8080/count")
	http.ListenAndServe(":8080", nil)
}

/*
	summary

	หัวใจสำคัญ: การเลือกใช้ `http.Handle` กับ struct เป็นรูปแบบที่ทรงพลังและยืดหยุ่น
	สำหรับการสร้างเว็บแอปพลิเคชันที่ซับซ้อนขึ้น

	1. http.Handle และ Handler ที่มี State (Stateful Handlers):
	   - `http.Handle` เหมาะอย่างยิ่งเมื่อ Handler ของเราต้องการ "จำ" หรือ "เก็บ" ข้อมูล (State) ไว้ระหว่าง request
	   - ตัวอย่าง: `CounterHandler` เก็บค่า `counter` ที่เพิ่มขึ้นทุกครั้งที่มีการเรียกเข้ามา
	   - การใช้ struct (`CounterHandler`) ทำให้เราสามารถมี field (`counter`, `mu`) สำหรับเก็บข้อมูลเหล่านี้ได้

	2. การจัดการ Concurrency (Goroutine Safety):
	   - เว็บเซิร์ฟเวอร์ใน Go จะจัดการแต่ละ request ใน Goroutine ของตัวเอง ซึ่งหมายความว่า handler ของเราอาจถูกเรียกใช้พร้อมกันหลายๆ ครั้ง
	   - หาก handler มีการแก้ไข state (เช่น `h.counter++`) เราจำเป็นต้องป้องกัน "Race Condition"
	   - `sync.Mutex` คือเครื่องมือสำคัญที่ใช้ในการ "Lock" เพื่อให้แน่ใจว่าในขณะใดขณะหนึ่ง จะมีเพียง Goroutine เดียวเท่านั้นที่สามารถเข้าถึงและแก้ไข state ได้
	   - ในตัวอย่าง: `h.mu.Lock()` และ `h.mu.Unlock()` จะครอบส่วนที่แก้ไข `counter` ไว้

	3. Dependency Injection:
	   - Struct handler เป็นวิธีที่ยอดเยี่ยมในการทำ Dependency Injection (การส่งต่อสิ่งที่ handler ต้องพึ่งพา)
	   - เราสามารถเตรียม dependency (เช่น database connection) ไว้ตอนสร้าง handler และส่งต่อเข้าไปใน struct ได้
	   - ตัวอย่างแนวคิด (ไม่ได้รันในโค้ดนี้):
	     // import "database/sql"
	     type AppHandler struct {
	         db *sql.DB // เก็บ database connection
	     }
	     func (h *AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	         // สามารถใช้ h.db เพื่อ query ข้อมูลได้เลย
	     }
	     // ตอนสร้าง:
	     // db, _ := sql.Open(...)
	     // http.Handle("/items", &AppHandler{db: db})

	4. สรุปเปรียบเทียบ Handle vs. HandleFunc:
	   - `HandleFunc`: สะดวก รวดเร็ว เหมาะสำหรับ handler ง่ายๆ ที่ไม่มี state และไม่มี dependency
	   - `Handle`: ยืดหยุ่นและทรงพลัง เหมาะสำหรับ handler ที่ซับซ้อน มี state, มี dependency, และต้องการการจัดการ concurrency ที่ดี
*/
