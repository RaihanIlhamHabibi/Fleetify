🚀 Fleetify - Fleet Maintenance System

Sistem internal untuk mengelola pemeliharaan armada kendaraan.
Technical test Junior Fullstack — Go (Fiber + GORM), MySQL 8, Vanilla JS + Bootstrap 5.

📌 Fitur
Kode	Deskripsi
F-01	SA membuat laporan awal + estimasi part/jasa → PENDING_APPROVAL
F-02	Approval menyetujui laporan → APPROVED
F-03	SA menyelesaikan pekerjaan + foto bukti → COMPLETED
F-04	Riwayat laporan (Nama SA, Nomor Polisi, Status, Tanggal)
B-01	Export CSV native JavaScript
B-02	Webhook HTTP POST async (goroutine) saat APPROVED / COMPLETED
⚡ Quick Start (Docker)
docker-compose up --build

👉 Buka browser: http://localhost:8080

Data awal otomatis di-seed dari init.sql (Docker) atau fallback seeder Go jika database kosong.

🔐 Environment Variables

Salin .env.example ke .env (opsional untuk Docker).

Variable	Default	Keterangan
DB_HOST	mysql	Host MySQL
DB_PORT	3306	Port MySQL
DB_USER	fleetify	Username database
DB_PASSWORD	fleetify_secret	Password database
DB_NAME	fleetify	Nama database
APP_PORT	8080	Port aplikasi
UPLOAD_DIR	./uploads	Folder upload foto
WEBHOOK_URL	(kosong)	URL webhook bonus (POST JSON)
RUN_SEEDER	true	Jalankan seeder Go jika tabel kosong

Contoh webhook:

WEBHOOK_URL=https://webhook.site/your-uuid docker-compose up --build
👤 Akun Testing
Username	Role	User ID (X-User-ID)
advisor_sa	SA	1
manager_approval	APPROVAL	2

Login via dropdown Login Simulasi, header X-User-ID otomatis dikirim ke API.

🔄 Alur Uji
Login sebagai advisor_sa → Buat laporan
Login sebagai manager_approval → Approve laporan
Kembali ke advisor_sa → Upload bukti & selesaikan pekerjaan
Lihat Riwayat Laporan → Export CSV (bonus feature)
🔌 API Endpoints

Semua endpoint /api/reports* (kecuali health/master) membutuhkan header:

X-User-ID: <user_id>
Method	Endpoint	Role	Keterangan
GET	/api/health	-	Health check
GET	/api/users	-	Daftar user
GET	/api/vehicles	-	Master kendaraan
GET	/api/master-items	-	Master part/jasa
GET	/api/reports	SA, APPROVAL	Riwayat laporan
GET	/api/reports/:id	SA, APPROVAL	Detail laporan
POST	/api/reports	SA	Buat laporan
PATCH	/api/reports/:id/approve	APPROVAL	Approve laporan
PATCH	/api/reports/:id/complete	SA	Selesaikan laporan
📦 POST /api/reports
{
  "vehicle_id": 1,
  "odometer": 45000,
  "complaint": "Rem berbunyi",
  "items": [
    { "item_id": 1, "quantity": 2 }
  ]
}
Multipart Support
vehicle_id
odometer
complaint
items (JSON string)
initial_photo (file)

Status awal: PENDING_APPROVAL
price_snapshot diambil dari master_items.

🏗️ Struktur Proyek
├── backend/          # Go + Fiber + GORM (Repository Pattern)
├── frontend/         # Vanilla JS + Bootstrap 5
├── schema.sql        # Skema InnoDB + FK
├── init.sql          # Data seeder SQL (Docker)
├── docker-compose.yml
└── Dockerfile
💻 Development Lokal (tanpa Docker)
Setup MySQL 8
→ import schema.sql + init.sql
Jalankan backend:
cd backend
go run .
Akses aplikasi:

http://localhost:8080

✅ Checklist Pengumpulan
 Repository GitHub Public
 Minimal 5 commit history
 schema.sql / init.sql / seeder otomatis
 README lengkap
📄 Lisensi

Proyek technical test — Fleetify internal recruitment.
