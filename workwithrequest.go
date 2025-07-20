package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type course struct {
	CourseId    int    `json:"id"`
	CourseName  string `json:"name"`
	CoursePrice int    `json:"price"`
	Instructor  string `json:"instructor"`
}

var CourseList []course

func init() {
	CoursesJson := `[
		{
			"id": 1,
			"name": "Golang",
			"price": 100,
			"instructor": "John Doe"
		},
		{
			"id": 2,
			"name": "Python",
			"price": 200,
			"instructor": "Jane Smith"
		},
		{
			"id": 3,
			"name": "Java",
			"price": 150,
			"instructor": "Bob Johnson"
		}
	]`

	err := json.Unmarshal([]byte(CoursesJson), &CourseList)
	if err != nil {
		log.Fatal(err)
	}
}

func getNextId() int {
	highestId := -1
	for _, course := range CourseList {
		if course.CourseId > highestId {
			highestId = course.CourseId
		}
	}
	return highestId + 1
}

func courseHandler(w http.ResponseWriter, r *http.Request) {
	// Concurrency Note: This handler is not thread-safe because it modifies the global
	// CourseList slice. In a real-world application, a mutex (sync.Mutex) should be
	// used to protect access to CourseList, similar to the handler.go example.
	switch r.Method {
	case http.MethodGet:
		courseJson, err := json.Marshal(CourseList)
		if err != nil {
			log.Printf("Error marshaling courses: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(courseJson)

	case http.MethodPost:
		var newCourse course
		// Use io.ReadAll instead of the deprecated ioutil.ReadAll (since Go 1.16)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Cannot read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, &newCourse)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// The client should not be able to set the ID.
		// We can enforce this by checking if an ID was provided.
		if newCourse.CourseId != 0 {
			http.Error(w, "Course ID is auto-generated and should not be provided.", http.StatusBadRequest)
			return
		}

		newCourse.CourseId = getNextId()
		CourseList = append(CourseList, newCourse)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// It's a good practice to return the created resource in the response body.
		json.NewEncoder(w).Encode(newCourse)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/courses", courseHandler)
	http.ListenAndServe(":8080", nil)
	log.Println("Server is running on http://localhost:8080")
}

/*
	summary

	หัวใจสำคัญ: การสร้าง RESTful API พื้นฐานใน Go สำหรับการจัดการข้อมูล (CRUD - Create, Read)
	โดยใช้ `http.HandleFunc` และการจัดการข้อมูล JSON

	1. การจัดการ HTTP Methods ที่แตกต่างกัน:
	   - Handler function (`courseHandler`) สามารถตรวจสอบ method ของ request ที่เข้ามาได้จาก `r.Method`
	   - `switch r.Method` เป็นรูปแบบที่นิยมใช้เพื่อแยก logic การทำงานสำหรับแต่ละ method (เช่น GET, POST, PUT, DELETE)
	   - หากเจอ method ที่ไม่รองรับ ควรตอบกลับด้วย `http.StatusMethodNotAllowed`

	2. การทำงานกับ JSON (Encoding/Decoding):
	   - Go มี package `encoding/json` ที่ทรงพลังสำหรับการแปลงข้อมูล
	   - **Marshal (Encoding):** การแปลงข้อมูลจาก Go struct/slice ไปเป็น JSON byte array (`json.Marshal`) เพื่อใช้ในการส่ง response กลับไปให้ client
	   - **Unmarshal (Decoding):** การแปลงข้อมูลจาก JSON byte array (ที่อ่านมาจาก request body) มาเป็น Go struct (`json.Unmarshal`) เพื่อนำข้อมูลไปใช้งานต่อ
	   - Struct tags (`json:"..."`) ใช้สำหรับ map ชื่อ field ใน Go struct กับ key ใน JSON

	3. การอ่านข้อมูลจาก Request Body:
	   - สำหรับ request ที่มี body (เช่น POST, PUT), เราสามารถอ่านข้อมูลได้จาก `r.Body`
	   - `ioutil.ReadAll(r.Body)` เป็นวิธีที่ง่ายในการอ่านข้อมูลทั้งหมดใน body ออกมาเป็น byte slice
	   - (หมายเหตุ: ใน Go 1.16+ แนะนำให้ใช้ `io.ReadAll` แทน `ioutil.ReadAll`)

	4. การส่ง Response กลับไปยัง Client:
	   - `w.Header().Set("Content-Type", "application/json")`: เป็นการบอก client ว่าข้อมูลที่ส่งกลับไปเป็นรูปแบบ JSON
	   - `w.WriteHeader(http.StatusOK)`: ใช้กำหนด HTTP Status Code เพื่อบอกผลลัพธ์ของการทำงาน (เช่น 200 OK, 201 Created, 400 Bad Request)
	   - `w.Write(...)`: ใช้สำหรับเขียน body ของ response

	5. การจัดการ State (In-Memory Database):
	   - ในตัวอย่างนี้ เราใช้ Global Variable (`CourseList`) เพื่อจำลองการเก็บข้อมูลในหน่วยความจำ (In-memory)
	   - `init()` function จะถูกเรียกทำงานเพียงครั้งเดียวก่อน `main()` เหมาะสำหรับการเตรียมข้อมูลเริ่มต้น
	   - **ข้อควรระวัง:** การใช้ Global Variable ในลักษณะนี้ **ไม่ปลอดภัยสำหรับการทำงานพร้อมกัน (Not Concurrency-Safe)** หากมีหลาย request เข้ามาแก้ไข `CourseList` พร้อมกัน อาจเกิด Race Condition ได้ ควรใช้ Mutex (`sync.Mutex`) เพื่อป้องกันปัญหานี้ (เหมือนในตัวอย่าง `handler.go`)
*/
