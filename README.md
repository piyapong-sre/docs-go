# Go ภาษาไทย

เว็บ docs สอนเขียนโปรแกรมภาษา Go ตั้งแต่พื้นฐานจนถึงขั้นสูง เนื้อหาภาษาไทย (ศัพท์เทคนิคภาษาอังกฤษ) พร้อมตัวอย่าง code ที่รันได้จริง และบทสร้าง Web API แบบ best practice

สร้างด้วย [Astro Starlight](https://starlight.astro.build)

## เนื้อหา

| Part | โฟลเดอร์ | หัวข้อ |
|---|---|---|
| 1. เริ่มต้น | `src/content/docs/start/` | Go คืออะไร, ติดตั้ง, Hello World, modules & toolchain |
| 2. พื้นฐาน | `src/content/docs/basics/` | ตัวแปร, control flow, function, slice, map, string/rune (UTF-8 ภาษาไทย), struct, pointer |
| 3. ระดับกลาง | `src/content/docs/intermediate/` | method, interface, error, defer/panic, package, generics, iterators, testing |
| 4. Concurrency | `src/content/docs/concurrency/` | goroutine, channel, select, sync, context, patterns, race detector |
| 5. ขั้นสูง | `src/content/docs/advanced/` | project layout, slog, embed, build, profiling, memory, reflection, code generation, tooling |
| 6. สร้าง Web API | `src/content/docs/web-api/` | REST API ด้วย `net/http` + PostgreSQL + JWT ตั้งแต่ต้นจน deploy |

ตัวอย่าง code:

- `examples/learn/` — Go module ของ snippet ทุกบท (part 1–5) ถูก import เข้าหน้า docs โดยตรง
- `examples/tasks-api/` — REST API ฉบับเต็มที่ใช้ใน part 6 (ดู [README](examples/tasks-api/README.md))

## รันเว็บ docs

ต้องมี Node.js 22+ และ pnpm

```bash
pnpm install
pnpm dev        # http://localhost:4321
pnpm build      # build static site ไปที่ dist/
pnpm preview    # ดูผล build
```

`dist/` เป็น static site ล้วน วางบน S3, Cloudflare Pages, GitHub Pages, Netlify ฯลฯ ได้ทันที

## ตรวจตัวอย่าง code

ต้องมี Go 1.26+

```bash
cd examples/learn && go vet ./... && go test -race ./...
cd examples/tasks-api && go vet ./... && go test -race ./...
```
