
# Authentication Microservice 

เป็นระบบยืนยันตัวตนที่ถูกพัฒนาขึ้นเพื่อให้บริการเกี่ยวกับการจัดการผู้ใช้และการยืนยันตัวตน โดยสร้างเป็น microservice แยกต่างหากจากส่วนอื่นๆ ของระบบ ช่วยให้สามารถนำไปใช้ซ้ำในโปรเจคอื่นๆ และสามารถ scale ได้อย่างอิสระ

<br><br>
## Tech Stack

**Go (Golang)  :**  ภาษาหลักที่ใช้ในการพัฒนา เนื่องจากมีประสิทธิภาพสูง, รองรับ concurrent programming, compile เป็น binary ขนาดเล็ก และเหมาะสำหรับการพัฒนา microservices

**JWT (JSON Web Tokens) :** ใช้ JWT สำหรับการยืนยันตัวตนเพราะ มีความปลอดภัยด้วยการเข้ารหัส เป็นมาตรฐานที่ได้รับการยอมรับอย่างกว้างขวางในอุตสาหกรรม

**gRPC :**  ใช้สำหรับการสื่อสารระหว่าง services เพราะ มีประสิทธิภาพสูงกว่า REST API ด้วย Protocol Buffers

**Docker :** รองรับการทำงานแบบ microservices ได้อย่างเหมาะสม ช่วยแยกส่วนของแอปพลิเคชันออกจากกัน ทำให้การดูแลรักษาง่ายขึ้น

<br><br>
## API Document

| API Endpoint | Method | Description | Request Body | Response |
|-------------|--------|-------------|--------------|----------|
| `/auth/register` | POST | ลงทะเบียนผู้ใช้ใหม่ | `{"email": "user@example.com", "password": "password123", "first_name": "John", "last_name": "Doe"}` | `{"success": true, "message": "Registration successful", "user_id": "1"}` |
| `/auth/login` | POST | เข้าสู่ระบบและรับ JWT token | `{"email": "user@example.com", "password": "password123"}` | `{"success": true, "message": "Login successful", "token": "eyJhbGc..."}` |
| `/auth/logout` | POST | ออกจากระบบ (ทำให้ token ไม่สามารถใช้งานได้) | *ไม่มี* | `{"success": true, "message": "Logout successful"}` |
| `/users` | GET | ดึงรายการผู้ใช้ (พร้อมการกรอง) | Query params: `name`, `email`, `page`, `page_size` | `{"users": [...], "pagination": {"total": 25, "page": 1, "page_size": 10, "total_pages": 3}}` |
| `/users/:id` | GET | ดึงข้อมูลผู้ใช้ตาม ID | *ไม่มี* | `{"profile": {"id": "1", "email": "user@example.com", "first_name": "John", "last_name": "Doe", ...}}` |
| `/users/:id` | PUT | อัปเดตข้อมูลผู้ใช้ | `{"first_name": "New Name", "email": "new@example.com", ...}` | `{"user": {"id": "1", "email": "new@example.com", "first_name": "New Name", ...}}` |
| `/users/:id` | DELETE | ลบผู้ใช้ (Soft Delete) | *ไม่มี* | `{"message": "User deleted successfully"}` |

<br><br>
## [My Work Process](https://docs.google.com/document/d/1swFCn2uYX76xDOyTOceVm7SzhA1RU5OF-YCCSork-44/edit?usp=sharing)

1. ฉันเริ่มโปรเจคโดยใช้ Claude AI ด้วย prompt ที่ระบุความต้องการของโปรเจค เพราะต้องการเรียนรู้วิธีการพัฒนา microservice ด้วย Go และต้องการเห็นแนวทางการพัฒนาที่เป็นมาตรฐานในอุตสาหกรรม Claude ช่วยให้ฉันเข้าใจโครงสร้างโปรเจคและวิธีการ implement API ต่างๆ ได้อย่างรวดเร็ว


2. การ Setup โปรเจค สร้างโครงสร้างโปรเจคตามหลักการทำ microservice และ ตั้งชื่อ file-folder ตาม best practice ที่ go แนะนำ

3. ติดตั้ง dependencies ที่จำเป็น

4. การ Implement และ Flow ของโค้ด

```
Register API Flow

1. ผู้ใช้ส่งข้อมูลการลงทะเบียน (email, password, ชื่อ, นามสกุล) ผ่าน POST request
2. handler.Register รับข้อมูลและตรวจสอบความถูกต้อง
3. เรียก service.Register เพื่อประมวลผลข้อมูล
4. ตรวจสอบว่า email ไม่ซ้ำกับที่มีอยู่แล้ว
5. เข้ารหัสรหัสผ่านด้วย bcrypt
6. บันทึกข้อมูลผู้ใช้ลงในฐานข้อมูล
7. ส่งผลลัพธ์กลับไปยังผู้ใช้
```
![Register](./images/register.png)

```
Login API Flow

1. ผู้ใช้ส่งข้อมูลการเข้าสู่ระบบ (email, password) ผ่าน POST request
2. handler.Login รับข้อมูลและเรียก service.Login
3. ค้นหาผู้ใช้จาก email ในฐานข้อมูล
4. ตรวจสอบรหัสผ่านด้วย bcrypt
5. หากข้อมูลถูกต้อง สร้าง JWT token 
6. ที่มีข้อมูลผู้ใช้และกำหนดเวลาหมดอายุ
7. ส่ง token กลับไปยังผู้ใช้
```

```
Logout API Flow

1. ผู้ใช้ส่ง token ผ่าน Authorization header
2. authMiddleware ตรวจสอบความถูกต้องของ token
3. handler.Logout เรียก service.Logout
4. เพิ่ม token ปัจจุบันเข้าไปใน blacklist หรือยกเลิกการใช้งาน
5. ส่งผลลัพธ์การออกจากระบบสำเร็จ
```

```
User-List API Flow

1. ผู้ใช้ส่ง GET request พร้อม query parameters สำหรับการกรอง (name, email) และ pagination (page, page_size)
2. authMiddleware ตรวจสอบสิทธิ์การเข้าถึง
3. handler.ListUsers สกัด parameters และเรียก service.
4. ListUsers สร้าง query ตามเงื่อนไขการกรอง
5. ดึงข้อมูลผู้ใช้ตาม pagination
6. คำนวณข้อมูล pagination และส่งกลับผู้ใช้
```

```
User-Profile API Flow

1. ผู้ใช้ส่ง GET request พร้อม user ID เป็น path parameter
2. authMiddleware ตรวจสอบสิทธิ์การเข้าถึง
3. handler.GetUserProfile ดึง user ID และเรียก service.
4. GetUserProfile
5. ตรวจสอบว่าผู้ใช้มีสิทธิ์ดูข้อมูลนี้หรือไม่(ต้องเป็นเจ้าของหรือ admin)
6. ดึงข้อมูลผู้ใช้จากฐานข้อมูลและส่งกลับ
```

```
User-Update API Flow

1. ผู้ใช้ส่ง PUT request พร้อม user ID และข้อมูลที่ต้องการอัปเดต
2. middleware.UpdateUserMiddleware ตรวจสอบความถูกต้องของข้อมูล
3. handler.UpdateUser ตรวจสอบสิทธิ์และเรียก service.UpdateUser
4. ตรวจสอบว่าผู้ใช้มีสิทธิ์อัปเดตข้อมูลนี้หรือไม่
5. ตรวจสอบความถูกต้องของข้อมูล (เช่น รูปแบบอีเมล, ความยาวรหัสผ่าน)
6. อัปเดตข้อมูลในฐานข้อมูลและส่งข้อมูลที่อัปเดตแล้วกลับไป
```

```
User-Delete API Flow (Soft Delete)

1. ผู้ใช้ส่ง DELETE request พร้อม user ID
2. authMiddleware ตรวจสอบว่าผู้ใช้เป็น admin หรือไม่
3. handler.DeleteUser เรียก service.DeleteUser
4. แทนที่จะลบข้อมูลออกจากฐานข้อมูลจริงๆ ระบบจะกำหนดค่า 
5. deleted_at เป็นเวลาปัจจุบัน
6. ส่งผลลัพธ์การลบสำเร็จกลับไป
```

<br><br>
## What I learned in this project ?
สิ่งที่ได้เรียนรู้ในโปรเจคนี้

**Microservice Architecture :** ได้รู้จัก microservice ว่าคือสถาปัตยกรรมการพัฒนาซอฟต์แวร์ที่แบ่งแอปพลิเคชันออกเป็นบริการย่อยๆ ที่เป็นอิสระต่อกัน แต่ละบริการทำงานเฉพาะด้านและสื่อสารกันผ่าน API ช่วยให้พัฒนา ทดสอบ และ deploy ได้ง่ายขึ้น สามารถ scale แต่ละส่วนได้อย่างอิสระ และทำงานร่วมกับเทคโนโลยีที่หลากหลายได้

**gRPC :** ได้เรียนรู้เกี่ยวกับ gRPC ซึ่งเป็น framework สำหรับการสื่อสารระหว่าง services ที่มีประสิทธิภาพสูงกว่า REST API แบบดั้งเดิม ทำงานบน HTTP/2 และใช้ Protocol Buffers สำหรับการ serialize ข้อมูล ช่วยให้การพัฒนา API เป็นไปอย่างมีประสิทธิภาพและมีความเป็นมาตรฐาน

**Go Programming Language :** ได้เรียนรู้ภาษา Go ซึ่งเป็นภาษาที่มีประสิทธิภาพสูง มีความเรียบง่าย และเหมาะสำหรับการพัฒนาแอปพลิเคชันเว็บและ microservices ได้เรียนรู้เกี่ยวกับ goroutines, channels, error handling และการใช้งาน packages ต่างๆ


<br><br>
**Authentication และ Authorization :** เข้าใจความแตกต่างระหว่าง Authentication (การยืนยันตัวตน) และ Authorization (การตรวจสอบสิทธิ์) มากขึ้น Authentication คือการตรวจสอบว่าผู้ใช้เป็นใคร โดยใช้ข้อมูลเช่น email และ password ส่วน Authorization คือการตรวจสอบว่าผู้ใช้มีสิทธิ์ทำอะไรได้บ้าง  
ได้เรียนรู้การใช้ JWT (JSON Web Tokens) สำหรับการจัดการ session แบบ stateless ซึ่งช่วยให้ระบบ scale ได้ง่ายและไม่ต้องเก็บข้อมูล session ในฐานข้อมูล

## If you want to try running this project on your machine, how do you do it?
ถ้าอยากลองเอา project นี้ไปรันบนเครื่องของคุณ ให้ทำตามนี้ 
1. ติดตั้ง Prerequisites:
2. ติดตั้ง Go (version 1.18 หรือสูงกว่า)
3. ติดตั้ง Docker และ Docker Compose
4. ติดตั้ง Postman (สำหรับทดสอบ API)
