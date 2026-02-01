
# PRD – WhatsApp Finance Bot (GOWA + Go + AI)

> **Versi: Detailed / Engineering-Ready PRD**  
> Dokumen ini menggabungkan **PRD awal yang detail** + **seluruh tambahan saran arsitektural, AI, dan reliability**.  
> Ditujukan untuk: manusia (engineer/product) **dan** AI agent (Claude).

---

## 0. Ringkasan Produk

**Nama Produk:** WhatsApp Finance Bot  
**Channel:** WhatsApp (GOWA Engine)  
**Backend:** Go  
**Database:** PostgreSQL (VPS existing)  
**AI:** OpenAI API – model **4o-mini** (text & vision)  

**Fungsi Utama:**
Bot WhatsApp yang memungkinkan siapa pun mencatat keuangan pribadi (pemasukan & pengeluaran) melalui **bahasa natural** (teks atau gambar), menyimpannya secara terstruktur, dan menampilkan rekap keuangan fleksibel.

**Monetisasi:**
- Free user: maksimal **10 transaksi pencatatan**
- Premium user: **Rp10.000 / bulan** (unlimited)
- Upgrade dilakukan manual via admin WhatsApp **081389592985**

---

## 1. Problem Statement

Banyak orang ingin mencatat keuangan pribadi, namun:
- Tidak konsisten membuka aplikasi
- Pencatatan manual terasa ribet
- Aplikasi keuangan terlalu kompleks

Solusi yang diinginkan adalah **pencatatan secepat chat**, tanpa perlu struktur kaku dari user.

---

## 2. Goals & Success Metrics

### Goals
1. Membuat pencatatan keuangan semudah mengirim chat.
2. Memungkinkan input bebas (natural language & gambar).
3. Menyediakan rekap yang fleksibel dan mudah dipahami.
4. Menjaga kualitas data meskipun input tidak terstruktur.

### Success Metrics (KPI)
- ≥ 85% pesan pencatatan berhasil diproses tanpa koreksi manual
- Akurasi klasifikasi transaksi ≥ 90%
- Free limit enforcement 100%
- Error AI yang sampai ke user < 5%

---

## 3. Target Users & Personas

### User Umum
- Individu yang ingin mencatat keuangan pribadi
- Tidak ingin belajar format khusus
- Menggunakan bahasa sehari-hari

### Admin
- Mengelola upgrade/downgrade user
- Memantau error dan kasus khusus

---

## 4. High-Level User Journey

1. User mengirim pesan ke bot WhatsApp
2. Jika nomor belum terdaftar → onboarding
3. User mencatat transaksi via teks atau gambar
4. Bot menyimpan transaksi terstruktur
5. User meminta rekap atau edit data

---

## 5. Onboarding & Conversation State

### 5.1 User Baru

Jika nomor belum ada di database:
- Bot **SELALU** meminta user memilih paket

Pesan onboarding (diulang sampai dipilih):
> "Halo! Aku bot pencatat keuangan 📒\nPilih paket:\n1️⃣ Free (10 transaksi)\n2️⃣ Premium Rp10rb/bulan (hubungi admin 081389592985)"

Jika user memilih:
- **Free** → plan=FREE, free_tx_count=0
- **Premium** → plan=PENDING_PREMIUM (tetap bisa pakai free)

### 5.2 Conversation State (WAJIB)

Setiap user memiliki state:
- NEW_USER
- ONBOARDING_SELECT_PLAN
- ACTIVE
- AWAITING_CONFIRM_RECORD
- EDITING_TRANSACTION
- ERROR_STATE

State disimpan di DB (`conversation_state`) dan memiliki expiry.

---

## 6. Pencatatan Transaksi via Teks

### 6.1 Trigger

Pesan dengan:
- Keyword eksplisit: "catat"
- Atau hasil AI intent detection (fase lanjut)

Contoh valid:
- "catat pemasukan 10000 gaji"
- "aku dapat uang dari jual motor 20 juta"
- "beli bensin 50rb"

Jika intent ambigu → AI menentukan confidence.

### 6.2 Flow

1. Pesan dikirim ke AI parser
2. AI mengembalikan JSON terstruktur
3. Sistem mengecek confidence:
   - ≥ 0.7 → auto simpan
   - 0.4–0.69 → minta klarifikasi
   - < 0.4 → tolak + contoh
4. Transaksi disimpan
5. Bot mengirim konfirmasi + TX_ID
6. User bisa `undo` dalam 60 detik

### 6.3 Free Limit Rule

- Setiap transaksi increment `free_tx_count`
- Jika count > 10 → tolak & arahkan upgrade

---

## 7. Pencatatan via Gambar

### Flow
1. User mengirim gambar (struk, transfer, dsb)
2. Bot download & kirim ke AI vision
3. AI mengekstrak nominal, tanggal, konteks
4. Jika data tidak jelas → minta klarifikasi
5. Transaksi disimpan sebagai EXPENSE/INCOME

Jika gagal total:
> "Aku belum bisa membaca gambar ini 😅 Bisa kirim ulang atau ketik manual?"

---

## 8. Edit, Delete & Undo

### Undo
- Berlaku 60 detik setelah simpan
- Undo tercatat di audit log

### Edit
- "edit transaksi terakhir jadi 15000"
- "edit TX#123 nominal 20000"

### Delete
- "hapus TX#123"

Semua perubahan **WAJIB** dicatat di audit log.

---

## 9. Rekap & Laporan

### Supported Queries
- Rekap harian / kemarin
- Rekap mingguan
- Rekap bulanan
- Rekap rentang tanggal ("1–3 Mei")

### Output
- Total pemasukan
- Total pengeluaran
- Net balance
- Top kategori
- Insight ringan (opsional)

---

## 10. Admin Flow

**Admin allowlist:** 081389592985

### Upgrade
Format:
`upgrade <msisdn> monthly <dd/mm>`

Efek:
- plan=PREMIUM
- premium_until = start_date + 1 bulan

### Command Lain
- status <msisdn>
- extend <msisdn> <n> month
- setfree <msisdn>
- block / unblock <msisdn>

Admin command bypass state machine.

---

## 11. AI Specification (Summary)

- Model: 4o-mini
- Output JSON strict
- Normalisasi angka (rb, jt)
- Timezone-aware date parsing
- Confidence threshold enforced
- `ai_version` disimpan di DB

---

## 12. Data Integrity & Reliability (WAJIB)

### Idempotency
- Dedup berdasarkan `wa_message_id`

### Audit Trail
- Semua create/edit/delete/undo tercatat

### Error Handling
- AI timeout → retry 2x
- DB error → rollback, quota tidak bertambah
- User error → pesan manusiawi

---

## 13. Non-Functional Requirements

- Latency AI < 12 detik
- Rate limit per user
- Logging lengkap (user_id, wa_message_id)
- Tidak menyimpan data sensitif di log

---

## 14. Acceptance Criteria (MVP)

1. User baru wajib memilih paket
2. Free user tidak bisa lebih dari 10 transaksi
3. Semua transaksi bisa di-edit & di-undo
4. Rekap sesuai bahasa natural
5. Admin upgrade otomatis set expiry

---

## 15. Future Enhancements

- Intent detection tanpa keyword
- Insight keuangan lanjutan
- Kategori custom per user
- Export data
- Web dashboard
